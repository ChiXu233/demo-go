package controller

import (
	. "demo-go/model"
	. "demo-go/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

// CreateOrUpdateDictController 创建or更新字典表
func CreateOrUpdateDictController(c *gin.Context) {
	var requestBody = &NewDictRequest{}
	err := c.BindJSON(&requestBody)
	if err != nil {
		SendServerErrorResponse(c, "参数解析错误", err)
		return
	}
	if requestBody.DictName == "" {
		SendParameterResponse(c, "字典名称不能为空", err)
		return
	}
	if requestBody.ProjectId == 0 {
		SendParameterResponse(c, "项目ID不能为空", err)
		return
	}

	//查询关联项目是否存在
	var project Project
	err = QueryEntity(requestBody.ProjectId, &project)
	if err != nil {
		SendServerErrorResponse(c, "查找关联项目失败", err)
		return
	}
	projectInfo := project.Info()

	//查找解析方案是否存在
	var parse Parse
	selector := make(map[string]interface{})
	selector["project_id"] = project.ID
	err = QueryEntityByFilter(&selector, &parse)
	if err != nil {
		SendServerErrorResponse(c, "查询解析方案失败", err)
		return
	}
	if parse.ID == 0 {
		SendServerErrorResponse(c, "解析方案不存在", err)
		return
	}

	//验证解析方案是否被关联
	var relateDict Dict
	selector = make(map[string]interface{})
	selector["parse_id"] = parse.ID
	err = QueryEntityByFilter(&selector, relateDict)
	if err != nil {
		SendServerErrorResponse(c, "查询字典表失败", err)
		return
	}
	dictName := requestBody.DictName

	//校验是否重名
	var existDict Dict
	selector = make(map[string]interface{})
	selector["dict_name"] = dictName
	err = QueryEntityByFilter(&selector, existDict)
	if err != nil {
		SendServerErrorResponse(c, "查找字典表失败", err)
		return
	}
	if requestBody.ID == 0 {
		//新增
		if relateDict.ID != 0 {
			SendServerErrorResponse(c, "解析方案已被关联", err)
			return
		}
		if existDict.ID > 0 {
			SendServerErrorResponse(c, "已经存在同名字典", err)
			return
		}

		//校验字典表是否存在
		var dictList []Dict
		selector = make(map[string]interface{})
		selector["dict_name"] = dictName
		selector["project_info"] = project.Info()
		selector["parse_info"] = parse.ParseName
		err = QueryList(&selector, &dictList)
		if err != nil {
			SendServerErrorResponse(c, "读取字典列表失败", err)
			return
		} else if len(dictList) > 0 {
			SendParameterResponse(c, "字典表("+dictName+")已经存在", err)
			return
		}
		//新增字典表
		var dict Dict
		dict.DictName = dictName
		dict.ProjectInfo = projectInfo
		dict.ParseInfo = parse.ParseName
		dict.ProjectID = project.ID
		dict.ParseID = parse.ID
		dict.ScanType = project.ScanType
		scanTypeConvert(&dict.ScanType)
		dict.Creator = "admin"
		dict.Comment = requestBody.Comment

		transaction := DB.Begin()
		err = CreateEntity(transaction, &dict)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "字典表数据库插入失败", err)
			return
		}
		operationLog := "新增字典表：字典表名称：" + dictName + ";"
		err = Log(c, "新增字典表", "标准化管理", operationLog, 2, transaction)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "日志生成失败", err)
			return
		}
		transaction.Commit()
		SendNormalResponse(c, dict)
		return
	} else {
		//更新字典表
		if existDict.ID > 0 && existDict.ID != requestBody.ID {
			SendServerErrorResponse(c, "存在重名字典", nil)
			return
		}
		//更新字典表
		var dict Dict
		err = QueryEntity(requestBody.ID, &dict)
		if err != nil {
			SendServerErrorResponse(c, "读取字典表失败", err)
			return
		}
		//判断是否关联至其他字典表
		if relateDict.ID > 0 && relateDict.ID != dict.ID {
			SendServerErrorResponse(c, "该解析方案已关联至其他字典表", nil)
			return
		}
		transaction := DB.Begin()

		dict.DictName = dictName
		dict.ProjectInfo = projectInfo
		dict.ParseInfo = parse.ParseName
		dict.ProjectID = parse.ID
		dict.ScanType = ScanTypeMap[project.ScanType]
		dict.Comment = requestBody.Comment
		err = UpdateEntities(transaction, &dict)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新字典表失败", err)
			return
		}
		operationLog := "更新字典表:字典表ID:" + fmt.Sprintf("%d", dict.ID) + ";"
		err = Log(c, "更新字典表", "标准化管理", operationLog, 2, transaction)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "记录日志失败", err)
			return
		}
		transaction.Commit()
		SendNormalResponse(c, dict)
		return
	}
}

// QueryDictListController 查看字典表列表
func QueryDictListController(c *gin.Context) {
	var dictList []Dict
	selector := make(map[string]interface{})
	selector["order"] = "ID desc"

	dictName := c.Query("dict_name")
	if dictName != "" {
		selector["dict_name"] = dictName
	}
	projectIdStr := c.Query("project_id")
	if projectIdStr != "" {
		projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
		if projectId == 0 {
			SendServerErrorResponse(c, "读取projectID失败", err)
			return
		}
		selector["project_id"] = uint(projectId)
	}
	err := QueryList(&selector, &dictList)
	if err != nil {
		SendServerErrorResponse(c, "查询字典列表失败", err)
		return
	}
	SendNormalResponse(c, dictList)
}

// QueryDictController 查看字典表详情
func QueryDictController(c *gin.Context) {
	dictID, _ := strconv.Atoi(c.Query("dict_id"))
	categoryName := c.Query("category_name")
	codeName := c.Query("code_name")
	codeSource := c.QueryArray("code_source")
	var items []DictItem
	selector := make(map[string]interface{})
	selector["dict_id"] = uint(dictID)
	selector["order"] = "category_name,code"

	if categoryName != "" {
		selector["category_name"] = categoryName
	}
	if codeName != "" {
		selector["code_name"] = codeName
	}
	if len(codeSource) == 1 && codeSource[0] != "" {
		codeSourceInt, _ := strconv.Atoi(codeSource[0])
		if !(codeSourceInt == 1 || codeSourceInt == 2) {
			SendServerErrorResponse(c, "codeSource只能为1或2", nil)
			return
		}
		selector["code_source"] = codeSourceInt
	}
	err := QueryList(&selector, &items)
	if err != nil {
		SendServerErrorResponse(c, "查询dict失败", err)
		return
	}
	var returnList []DictItemList
	order := 1
	for _, item := range items {
		var returnDictItem DictItemList
		returnDictItem.BaseModel = item.BaseModel
		returnDictItem.BaseItem = item.BaseItem
		returnDictItem.Order = uint(order)
		returnDictItem.CodeSource = CodeSourceMap[item.CodeSource]

		order += 1
		returnList = append(returnList, returnDictItem)
	}
	if len(returnList) == 0 {
		returnList = []DictItemList{}
	}
	SendNormalResponse(c, returnList)

}
