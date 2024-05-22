package model

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OperationLog struct {
	BaseModel
	Name      string `gorm:"column:name" json:"name"`
	Account   string `gorm:"column:account" json:"account"`
	IP        string `gorm:"column:IP" json:"IP"`
	Operation string `gorm:"column:operation" json:"operation"`
	Module    string `gorm:"column:module" json:"module"`
	Detail    string `gorm:"column:detail" json:"detail"`
	Level     uint8  `gorm:"column:level" json:"level"`
}

func (OperationLog) TableName() string {
	return "operation_log"
}

func Log(c *gin.Context, operation string, module string, detail string, level uint8, DBExecutor ...*gorm.DB) error {
	var err error
	var log OperationLog
	var user User

	transaction := DB
	if len(DBExecutor) > 0 {
		transaction = DBExecutor[0]
	}
	accessTokenHeader := c.Request.Header["Access-Token"]
	if len(accessTokenHeader) < 1 {
		return errors.New("Access-Token为空")
	}
	accessToken := accessTokenHeader[0]
	err = GetUser(accessToken, &user)
	if err != nil {
		return err
	}
	log = OperationLog{
		Name:    user.Name,
		Account: user.Account,
		// 如果使用了nginx做反向代理，需要在nginx配置中加入
		//				  proxy_pass http://127.0.0.1:9000/api;
		//                proxy_set_header X-Real-IP $remote_addr;
		//                proxy_set_header X-Forward-For $remote_addr;
		IP:        c.ClientIP(),
		Operation: operation,
		Module:    module,
		Detail:    detail,
		Level:     level,
	}

	err = CreateEntity(transaction, &log)
	if err != nil {
		return err
	}
	return nil
}
