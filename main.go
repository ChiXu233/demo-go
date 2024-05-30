package main

import (
	"demo-go/config"
	"demo-go/controller"
	"demo-go/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	err = logger.SetLogger("log.json")
	if err != nil {
		panic(err)
	}
	err = model.InitDB()
	if err != nil {
		panic(err)
	}
	err = model.DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.UserRole{},
		&model.Token{},
		&model.OperationLog{},
		&model.Dict{},
	)
	if err != nil {
		panic(err)
	}
	transaction := model.DB.Begin()
	err = model.InitUser(transaction)
	if err != nil {
		transaction.Rollback()
		panic(err)
	}
	transaction.Commit()
	gin.SetMode("debug")
	//Logger中间件会将请求的方法、路径、状态码、处理时间等信息写入gin.DefaultWriter，即使在GIN_MODE=release模式下也会生效。默认情况下，gin.DefaultWriter = os.Stdout，也就是标准输出
	//Recovery中间件会从任何恐慌中恢复，并且如果有恐慌发生，它会写入一个500的错误
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	//添加路由
	v1 := r.Group("/api")
	{
		v1.GET("/ping", controller.Ping)
		//用户
		v1.POST("/user", controller.CreateOrUpdateUserController)
		v1.POST("/login", controller.LoginController)
		v1.GET("/token", controller.QueryUserController)
		v1.GET("/user", controller.GetUserController)
		//项目
		v1.POST("/project", controller.CreateOrUpdateProjectController) //新增或更新项目
		v1.GET("/projects", controller.GetProjectsController)           //查询项目信息
		v1.GET("/projects_info", controller.QueryProjectInfoController) //查找暂未被关联的项目

		//字典表
		v1.POST("/dict", controller.CreateOrUpdateDictController) //创建Or更新字典表
		v1.GET("/dict_list", controller.QueryDictListController)  //查看字典表列表
		v1.GET("/dict/:dict_id", controller.QueryDictController)  //查看字典表详情

		//预处理
		v1.GET("/preprocess", controller.QueryPreprocessListController)       //查询预处理列表
		v1.POST("/preprocess", controller.CreateOrUpdatePreprocessController) //新增或更新预处理规则

		//导入本地标准图(创建标准图组+读取标准图列表+修改standard_info、standard_item)
		v1.POST("importLocalStandard/", controller.ImportLocalStandardController) //导入本地标准图
	}

	// 捕捉不允许的方法
	r.NoMethod(controller.HandleNotFound)
	r.NoRoute(controller.HandleNotFound)
	err = r.Run(fmt.Sprintf("0.0.0.0:%d", config.Conf.APP.Port))
	if err != nil {
		panic(err)
	}
}

func Recover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": fmt.Sprintf("%v", r),
				"data":    nil,
			})
			c.Abort()
		}
	}()
	c.Next()
}
