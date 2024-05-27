package controller

import (
	. "demo-go/model"
	. "demo-go/utils"
	"github.com/gin-gonic/gin"
)

func QueryPreprocessListController(c *gin.Context) {
	//查询预处理规则列表
	var err error
	preprocessList := make([]PreProcess, 0)
	selector := make(map[string]interface{})
	selector["order"] = "id desc"
	err = QueryList(&selector, &preprocessList)
	if err != nil {
		SendServerErrorResponse(c, "查找列表失败", err)
		return
	}
	SendNormalResponse(c, preprocessList)
}

func CreateOrUpdatePreprocessController(c *gin.Context) {
	var err error
	var preprocess PreProcess
	err = c.BindJSON(&preprocess)
	if err != nil {
		SendServerErrorResponse(c, "", err)
		return
	}
	if preprocess.ProjectID == 0 || preprocess.ScanType == "" || preprocess.Name == "" {
		SendParameterResponse(c, "参数不全", err)
		return
	}
	if preprocess.ID == 0 {
		//新增
		var project Project
		err = QueryEntity(preprocess.ProjectID, &project)
		if err != nil {
			SendServerErrorResponse(c, "查找失败", err)
			return
		}

		//判断是否重名
		var preprocessQueryName PreProcess
		err = QueryEntityByFilter(&map[string]interface{}{"name": preprocess.Name}, preprocessQueryName, DB)
		if err != nil {
			SendServerErrorResponse(c, "", err)
			return
		}

		if preprocessQueryName.ID != 0 {
			SendParameterResponse(c, "名称重复", nil)
			return
		}

		//判断项目是否已有规则
		var preprocessQueryProject PreProcess
		err = QueryEntityByFilter(&map[string]interface{}{"project_id": preprocess.ProjectID, "scan_type": preprocess.ScanType}, &preprocessQueryProject)
		if err != nil {
			SendServerErrorResponse(c, "查找失败", err)
			return
		}
		if preprocessQueryProject.ID != 0 {
			SendParameterResponse(c, "项目此类目已存在预处理规则", err)
			return
		}
		err = CreateEntity(DB, &preprocess)
		if err != nil {
			SendServerErrorResponse(c, "新增预处理规则失败", err)
			return
		}
	} else {
		//更新
		//判断名称是否重复
		var preprocessQuery Project
		err = QueryEntityByFilter(&map[string]interface{}{"name": preprocess.Name}, &preprocessQuery, DB)
		if err != nil {
			SendServerErrorResponse(c, "", err)
			return
		}
		if preprocessQuery.ID != 0 && preprocessQuery.ID != preprocess.ID {
			SendParameterResponse(c, "名称重复", nil)
			return
		}
		//判断更新项目类目
		var preprocessQueryProject PreProcess
		err = QueryEntityByFilter(&map[string]interface{}{"project_id": preprocess.ProjectID, "scan_type": preprocess.ScanType}, &preprocessQueryProject)
		if err != nil {
			SendServerErrorResponse(c, "查找失败", err)
			return
		}
		if preprocessQueryProject.ID != 0 && preprocessQueryProject.ID != preprocess.ID {
			SendParameterResponse(c, "项目此类目已存在预处理规则", err)
			return
		}
		err = UpdateFields(DB, &PreProcess{}, &map[string]interface{}{"id": preprocess.ID},
			&map[string]interface{}{"name": preprocess.Name, "comment": preprocess.Comment, "project_id": preprocess.ProjectID, "scan_type": preprocess.ScanType})

		if err != nil {
			SendServerErrorResponse(c, "更新失败", err)
			return
		}

	}
	SendNormalResponse(c, "")
}
