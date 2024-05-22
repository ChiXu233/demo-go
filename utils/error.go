package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// HTTP Code
const (
	Normal         = 200 // 无错误
	ParameterError = 400 // 参数错误
	NotFound       = 404 // 无资源错误
	ServerError    = 500 // 系统错误
	UnknownError   = 503 // 未知错误
	Unauthorized   = 401 // 未授权
)

// APIException api错误的结构体
type APIException struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SendResponse(c *gin.Context, code int, message string, data interface{}, err error) {
	if err != nil {
		message = fmt.Sprintf("请求失败, %s %v", message, err)
	}
	c.JSON(code, APIException{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func SendParameterResponse(c *gin.Context, message string, err error) {
	SendResponse(c, ParameterError, message, nil, err)
}

func SendNormalResponse(c *gin.Context, data interface{}) {
	SendResponse(c, Normal, "请求成功", data, nil)
}

func SendNormalResponseMessageInfo(c *gin.Context, data string) {
	SendResponse(c, Normal, data, nil, nil)
}
func SendServerErrorResponse(c *gin.Context, message string, err error) {
	SendResponse(c, ServerError, message, nil, err)
}

func SendNotFoundResponse(c *gin.Context, message string) {
	SendResponse(c, NotFound, message, nil, nil)
}
