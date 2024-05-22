package controller

import (
	. "demo-go/model"
	. "demo-go/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type queryTokenResponse struct {
	Account string   `json:"account"`
	Roles   []string `json:"roles"`
}

func CreateOrUpdateUserController(c *gin.Context) {
	var requestBody User
	var err error
	err = c.BindJSON(&requestBody)
	if err != nil {
		SendParameterResponse(c, "读取请求参数错误", err)
		return
	}
	if requestBody.Account == "" || requestBody.Password == "" {
		SendServerErrorResponse(c, "账号/密码为空", err)
		return
	}
	if requestBody.ID == 0 {
		var temp User
		filter := make(map[string]interface{})
		filter["account"] = requestBody.Account
		if err = QueryEntityByFilter(&filter, &temp); err != nil {
			SendServerErrorResponse(c, "账号查找失败", err)
			return
		}
		if temp.Account != "" {
			SendServerErrorResponse(c, "账号重复", nil)
			return
		}
		rand.Seed(time.Now().UnixNano())
		//salt := strconv.FormatInt(rand.Int63(), 10)
		requestBody.Salt = strconv.FormatInt(rand.Int63(), 10)
		HashPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password+requestBody.Salt), bcrypt.DefaultCost)
		if err != nil {
			SendServerErrorResponse(c, "加密失败", err)
			return
		}
		requestBody.Password = string(HashPassword)
		transaction := DB.Begin()
		if err := CreateEntity(transaction, &requestBody); err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "创建用户失败", err)
			return
		}
		if err := Log(c, "创建用户", "用户管理", "创建用户: "+strconv.Itoa(int(requestBody.ID)), 2, transaction); err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "记录日志失败", err)
			return
		}
		transaction.Commit()
		SendNormalResponse(c, requestBody)
	} else {
		var userDB User
		//验证用户
		err := QueryEntity(requestBody.ID, &userDB)
		if err != nil {
			SendServerErrorResponse(c, "读取用户失败", err)
			return
		}
		userDB.Name = requestBody.Name
		userDB.Account = requestBody.Account
		if err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(requestBody.Password+requestBody.Salt)); err == nil {
			//未修改密码
			userDB.Password = requestBody.Password
			userDB.Salt = requestBody.Salt
		} else {
			//修改密码
			rand.Seed(time.Now().UnixNano())
			requestBody.Salt = strconv.Itoa(rand.Int())
			HashPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password+requestBody.Salt), bcrypt.DefaultCost)
			if err != nil {
				SendServerErrorResponse(c, "加密失败", err)
				return
			}
			userDB.Password = string(HashPassword)
			userDB.Salt = requestBody.Salt
		}
		transaction := DB.Begin()
		err = UpdateEntities(transaction, &userDB)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新用户失败", err)
			return
		}
		err = Log(c, "创建用户", "用户管理", "创建用户: "+strconv.Itoa(int(requestBody.ID)), 2, transaction)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "", err)
			return
		}
		transaction.Commit()
		SendNormalResponse(c, requestBody)
	}

}

func LoginController(c *gin.Context) {
	var user User
	var err error
	err = c.BindJSON(&user)
	if err != nil {
		SendParameterResponse(c, "读取请求参数错误", err)
		return
	}
	if user.Account == "" || user.Password == "" {
		SendServerErrorResponse(c, "账号/密码为空", err)
		return
	}
	err = GetUser(&user, nil)
	if err != nil {
		SendServerErrorResponse(c, "账号/密码错误", err)
		return
	}
	accessToken, err := GenerateToken(user.Account)
	if err != nil {
		SendServerErrorResponse(c, "生成token失败", err)
		return
	}
	err = CreateToken(DB, &user, accessToken)
	if err != nil {
		SendServerErrorResponse(c, "创建access_token失败", err)
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": accessToken, "data": map[string]interface{}{"access_token": accessToken},
	})
}

func QueryUserController(c *gin.Context) {
	var roles []string
	var user User
	accessTokenHeader := c.Request.Header["Access-Token"]
	if len(accessTokenHeader) < 1 {
		SendParameterResponse(c, "token为空", nil)
	}
	accessToken := accessTokenHeader[0]
	code, err := QueryUserAndRole(accessToken, &user, &roles)
	if err != nil {
		SendResponse(c, code, "Token校验失败", nil, err)
		return
	}
	c.JSON(200, queryTokenResponse{
		Account: user.Account,
		Roles:   roles,
	})
}

func QueryUserAndRole(accessToken string, user *User, roles *[]string) (int, error) {
	code, err := CheckToken(DB, accessToken, 0)
	if err != nil {
		return code, err
	}
	err = GetUser(accessToken, user)
	if err != nil {
		return 500, err
	}
	userRoles, _, err := QueryRoles(DB, user.ID)
	if err != nil {
		return 500, err
	}

	for _, userRole := range *userRoles {
		*roles = append(*roles, userRole.RoleName)
	}
	return 200, nil

}

func GetUserController(c *gin.Context) {
	var user []User
	query := DB
	query = query.Order("id desc")
	limitStr := c.Query("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 0
	}
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 0
	}
	if limit != 0 && page != 0 {
		query = query.Offset((page - 1) * limit).Limit(limit)
	}
	if err := query.Debug().Find(&user).Error; err != nil {
		SendServerErrorResponse(c, "查询项目失败", err)
		return
	}
	res := make(map[string]interface{})
	res["users"] = user
	SendNormalResponse(c, res)
}
