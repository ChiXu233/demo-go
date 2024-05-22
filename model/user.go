package model

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/wonderivan/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

const signal = "XuChi"

type User struct {
	gorm.Model
	Account  string `gorm:"column:account" json:"account"`
	Name     string `gorm:"column:name" json:"name"`
	Password string `gorm:"column:password" json:"password"`
	Salt     string `gorm:"column:salt" json:"salt"`
}

type Token struct {
	gorm.Model
	Account     string `gorm:"account"`
	AccessToken string `gorm:"access_token"`
	Available   *bool  `gorm:"available"`
	Expiration  int    `gorm:"expiration"`
	UserId      uint   `gorm:"user_id"`
}

type Role struct {
	gorm.Model
	Name       string `gorm:"column:name"`
	Permission int    `gorm:"column:permission"`
}

type UserRole struct {
	gorm.Model
	UserID   uint   `gorm:"user_id"`
	RoleID   uint   `gorm:"role_id"`
	Account  string `gorm:"account"`
	RoleName string `gorm:"role_name"`
}

type accessTokenInfo struct {
	jwt.StandardClaims
}

func (User) TableName() string {
	return "user"
}
func (Role) TableName() string {
	return "role"
}
func (Token) TableName() string {
	return "token"
}
func (UserRole) TableName() string {
	return "user_role"
}

// GetUser 查询user或校验token
func GetUser(info interface{}, user *User) error {
	var err error
	switch t := info.(type) {
	//传入user
	case *User:
		var userDB User
		query := DB.Where("account = ?", t.Account).First(&userDB)
		if err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(t.Password+userDB.Salt)); err != nil {
			return err
		}
		if err := query.Error; err != nil {
			return err
		}
		return nil
		//传入token
	case string:
		var token Token
		query := DB.Where("access_token = ?", t).First(&token)
		if err := query.Error; err != nil {
			return err
		}
		err := GetUser(token.UserId, user)
		if err != nil {
			return err
		}
		return nil
	case uint:
		err = QueryEntity(info.(uint), user)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid")
	}
}

func CreateToken(db *gorm.DB, user *User, token string) error {
	available := true
	accessToken := &Token{
		Account:     user.Account,
		AccessToken: token,
		Available:   &available,
		Expiration:  0,
		UserId:      user.ID,
	}
	if err := db.Create(accessToken).Error; err != nil {
		return err
	}
	return nil
}
func QueryRoles(db *gorm.DB, userID uint) (*[]UserRole, int, error) {
	var userRoles []UserRole
	query := db.Select("role_id, role_name").Where("user_id = ?", userID).Find(&userRoles)
	if err := query.Error; err != nil {
		return nil, 500, err
	}
	return &userRoles, 200, nil
}

func GetToken(db *gorm.DB, tokenStr string) (*Token, int, error) {
	var token = &Token{}
	query := db.Where("access_token = ?", tokenStr).Find(token)
	if err := query.Error; err != nil {
		return nil, 500, err
	}
	return token, 200, nil
}

func GetRole(db *gorm.DB, roleID uint) (*Role, int, error) {
	var role = &Role{}
	query := db.First(role, roleID)
	if err := query.Error; err != nil {
		return nil, 500, err
	}
	return role, 200, nil
}

// CheckToken 判断token是否存在、是否过期和是否有权限进行某操作
func CheckToken(db *gorm.DB, tokenStr string, operationPermission int) (int, error) {
	// 是否存在
	token, code, err := GetToken(db, tokenStr)
	if err != nil {
		return code, err
	}
	// 是否过期
	err = ParseToken(tokenStr)
	if err != nil {
		logger.Error(err)
		return 400, err
	}
	// 是否有权限
	if operationPermission <= 0 {
		return 200, nil
	}
	userRoles, code, err := QueryRoles(db, token.UserId)
	if err != nil {
		return code, err
	}
	var permission = 0
	for _, userRole := range *userRoles {
		role, code, err := GetRole(db, userRole.RoleID)
		if err != nil {
			return code, err
		}
		permission += role.Permission
	}
	if permission&operationPermission == 0 {
		return 401, errors.New("no permission for this operation")
	}
	return 200, nil
}

func GenerateToken(issuer string) (tokenString string, err error) {
	var expiration = 36000
	claims := &accessTokenInfo{
		jwt.StandardClaims{
			//设置过期时间
			ExpiresAt: time.Now().Add(time.Second * time.Duration(expiration)).Unix(),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(signal))
	return tokenString, err
}

func ParseToken(tokenSrt string) (err error) {
	_, err = jwt.Parse(tokenSrt, func(*jwt.Token) (interface{}, error) {
		return []byte(signal), nil
	})
	return
}

func InitUser(DBExecutor *gorm.DB) error {
	var err error
	roleUploader := Role{
		Name:       "matcher",
		Permission: 511,
	}
	roleReviewer := Role{
		Name:       "stander",
		Permission: 511,
	}
	userUploader := User{
		Account:  "uploader",
		Password: "1234561",
	}
	userReviewer := User{
		Account:  "reviewer",
		Password: "1234561",
	}
	result := DBExecutor.FirstOrCreate(&roleUploader, roleUploader)
	if err := result.Error; err != nil {
		return err
	}
	result = DBExecutor.FirstOrCreate(&roleReviewer, roleReviewer)
	if err = result.Error; err != nil {
		return err
	}
	result = DBExecutor.FirstOrCreate(&userUploader, userUploader)
	if err = result.Error; err != nil {
		return err
	}
	result = DBExecutor.FirstOrCreate(&userReviewer, userReviewer)
	if err = result.Error; err != nil {
		return err
	}
	userRoleUploader := UserRole{
		UserID:   userUploader.ID,
		RoleID:   roleUploader.ID,
		Account:  userUploader.Account,
		RoleName: roleUploader.Name,
	}
	userRoleReviewer := UserRole{
		UserID:   userReviewer.ID,
		RoleID:   roleReviewer.ID,
		Account:  userReviewer.Account,
		RoleName: roleReviewer.Name,
	}
	result = DBExecutor.FirstOrCreate(&userRoleUploader, userRoleUploader)
	if err = result.Error; err != nil {
		return err
	}
	result = DBExecutor.FirstOrCreate(&userRoleReviewer, userRoleReviewer)
	if err = result.Error; err != nil {
		return err
	}
	return nil
}
