package controller

import (
	"demo-go/config"
	. "demo-go/model"
	. "demo-go/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": config.Conf.DB.Name,
	})
}

// HandleNotFound 404处理
func HandleNotFound(c *gin.Context) {
	SendNotFoundResponse(c, "方法不允许")
}

// CreateOrUpdateProjectController 新建or更新
func CreateOrUpdateProjectController(c *gin.Context) {
	var requestBody = &NewProjectRequest{}
	err := c.BindJSON(requestBody)
	if err != nil {
		SendParameterResponse(c, "参数解析错误", err)
		return
	}
	if requestBody.CompanyName == "" {
		SendParameterResponse(c, "企业名称不可为空", nil)
		return
	}
	if requestBody.Name == "" {
		SendParameterResponse(c, "项目名称不可为空", nil)
		return
	}
	if requestBody.Location == "" {
		SendParameterResponse(c, "项目地点不可为空", nil)
		return
	}
	if len(requestBody.ScanType) == 0 {
		SendParameterResponse(c, "类目不可为空", nil)
		return
	}
	scanTypeConvert(&requestBody.ScanType)
	scanType := requestBody.ScanType
	companyName := requestBody.CompanyName
	name := requestBody.Name
	location := requestBody.Location
	comment := requestBody.Comment
	hdrProjectID := requestBody.HDRProjectID
	//查找项目是否存在
	var projectList []Project
	selector := make(map[string]interface{})
	selector["name"] = name
	selector["company_name"] = companyName
	err = QueryEntityByFilter(&selector, &projectList)
	if err != nil {
		SendServerErrorResponse(c, "读取项目列表失败", err)
		return
	}
	if requestBody.ProjectID == 0 {
		//新增操作
		if len(projectList) > 0 {
			//重名项目已经存在
			SendServerErrorResponse(c, "项目("+name+")已经存在", nil)
			return
		}
		//新增项目
		var project Project
		project.Name = name
		project.CompanyName = companyName
		project.ScanType = scanType
		project.Location = location
		project.Comment = comment
		project.HDRProjectID = hdrProjectID
		transaction := DB.Begin()
		err = CreateEntity(transaction, &project)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "创建项目-写入数据库失败", err)
			return
		}
		err = Log(c, "创建项目", "标准化管理", "创建项目: "+project.Info(), 2, transaction)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "写入日志失败", err)
			return
		}
		transaction.Commit()
		scanTypeConvert(&project.ScanType)
		SendNormalResponse(c, project)
		return
	} else {
		//更新操作
		//判断能否更新
		if len(projectList) > 0 {
			//重名项目已经存在
			for _, project := range projectList {
				if project.ID != requestBody.ProjectID {
					SendParameterResponse(c, "项目("+name+")已经存在", nil)
					return
				}
			}
		}
		var project Project
		//更新项目
		err := QueryEntity(requestBody.ProjectID, &project)
		if err != nil {
			SendServerErrorResponse(c, "读取项目失败", err)
			return
		}
		originProjectInfo := project.Info()
		project.CompanyName = companyName
		project.Name = name
		project.Location = location
		project.ScanType = scanType
		project.Comment = comment
		project.HDRProjectID = hdrProjectID
		transaction := DB.Begin()
		if err = UpdateEntity(transaction, &project); err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新项目失败", err)
			return
		}
		//更新与项目相关联的解析方案、字典表、匹配规则、标准图、数据组
		selector = make(map[string]interface{})
		selector["project_id"] = project.ID
		fields := make(map[string]interface{})
		fields["project_info"] = project.Info()
		err = UpdateFields(transaction, &Parse{}, &selector, &fields)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新解析方案失败", err)
			return
		}
		err = UpdateFields(transaction, &Dict{}, &selector, &fields)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新关联字典表失败", err)
			return
		}
		err = UpdateFields(transaction, &MatchRule{}, &selector, &fields)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新匹配规则失败", err)
			return
		}
		//err  = UpdateFields(transaction,&sourceData{},&selector,&fields)
		//if err !=nil{
		//	transaction.Rollback()
		//	SendServerErrorResponse(c,"更新关联字典表失败",err)
		//	return
		//}
		fields = make(map[string]interface{})
		fields["project"] = project.Info()
		err = UpdateFields(transaction, StandardGroup{}, &selector, &fields)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新标准图组失败", err)
			return
		}
		operationLog := fmt.Sprintf("%s->%s", originProjectInfo, project.Info())
		err = Log(c, "更新项目", "标准化管理", "更新项目："+operationLog, 2, transaction)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "更新失败", err)
			return
		}
		transaction.Commit()
		scanTypeConvert(&project.ScanType)
		SendNormalResponse(c, project)
	}
}

//CompanyName    string `gorm:"column:company_name" json:"company_name"`
//Name           string `gorm:"column:name" json:"name"`
//Location       string `gorm:"column:location" json:"location"`
//ScanType       string `gorm:"column:scan_type" json:"scan_type"`
//Comment        string `gorm:"column:comment" json:"comment"`
//HDRProjectID   uint   `gorm:"column:hdr_project_id" json:"hdr_project_id"`
//HDRProjectInfo string `gorm:"column:hdr_project_info" json:"hdr_project_info"

// GetProjectsController 获取所有项目信息
func GetProjectsController(c *gin.Context) {
	var projectList []Project
	// 数据检索
	company := c.Query("project_source")
	name := c.Query("project_name")
	location := c.Query("project_location")
	scanType := c.Query("scan_type")
	query := DB
	query2 := DB
	query = query.Order("id desc")
	query2 = query2.Order("id desc")
	if company != "" {
		query = query.Where("company_name = ?", company)
		query2 = query2.Where("company_name = ?", company)
	}
	if name != "" {
		query = query.Where("name = ?", name)
		query2 = query2.Where("name = ?", name)
	}
	if location != "" {
		query = query.Where("location = ?", location)
		query2 = query2.Where("location = ?", location)
	}
	if scanType != "" {
		types := strings.Split(scanType, "、")
		typesInterface := make([]interface{}, len(types))
		q := "scan_type like ?"
		typesInterface[0] = "%" + types[0] + "%"
		for i := 1; i < len(types); i++ {
			typesInterface[i] = "%" + types[i] + "%"
			q = q + "or" + "scan_type like ?"
		}
		query = query.Where(q, typesInterface...)
		query2 = query2.Where(q, typesInterface...)
	}
	var totalID []uint
	if err := query2.Model(&Project{}).Select("id").Find(&totalID).Error; err != nil {
		SendServerErrorResponse(c, "查询项目id失败", err)
		return
	}
	//分页
	limitStr := c.Query("limit")
	pageSize := c.Query("pageSize")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 0
	}
	page, err := strconv.Atoi(pageSize)
	if err != nil {
		page = 0
	}
	if limit != 0 && page != 0 {
		query = query.Offset((page - 1) * limit).Limit(limit)
	}
	if err := query.Debug().Find(&projectList).Error; err != nil {
		SendServerErrorResponse(c, "查询项目失败", err)
		return
	}
	for i := range projectList {
		scanTypeConvert(&projectList[i].ScanType)
	}
	res := make(map[string]interface{})
	res["projects"] = projectList
	res["total_id"] = totalID
	res["total"] = len(totalID)
	SendNormalResponse(c, res)
}

func scanTypeConvert(scanType *string) {
	scanTypeList := strings.Split(*scanType, "、")
	for i, s := range scanTypeList {
		scanTypeList[i] = ConvertMapping[s]
	}
	*scanType = strings.Join(scanTypeList, "、")
}

func QueryProjectInfoController(c *gin.Context) {
	//先查找所有已关联项目
	var projects []Project
	var UnusedProject []Project
	var standardGroupList []StandardGroup
	Rename := make(map[string]bool)
	selector := make(map[string]interface{})
	query := c.Query("query")
	//Table := c.Query("table")
	if query == "" {
		//query为空则为检索所有可关联项目
		err := QueryList(&selector, &projects)
		if err != nil {
			SendServerErrorResponse(c, "查询项目失败", err)
			return
		}
		err = QueryList(&selector, &standardGroupList)
		if err != nil {
			SendServerErrorResponse(c, "查询标准图组失败", err)
			return
		}
		for k := range projects {
			//列出所有项目
			TypeStrArr := strings.Split(projects[k].ScanType, "、")
			for _, v := range standardGroupList {
				//对已经关联完毕的数据进行匹配
				if v.ProjectID == projects[k].ID {
					//对已经关联的去除
					for i := 0; i < len(TypeStrArr); i++ {
						if TypeStrArr[i] == v.ScanType {
							TypeStrArr = append(TypeStrArr[:i], TypeStrArr[i+1:]...)
						}
					}
					//防止数组只有一个元素从而去除失败
					if len(TypeStrArr) == 1 && TypeStrArr[0] == v.ScanType {
						TypeStrArr = []string{}
					}
				}
			}

			projects[k].ScanType = strings.Join(TypeStrArr, "、")
			if projects[k].ScanType != "" {
				//对companyName进行去重操作
				if _, ok := Rename[projects[k].CompanyName]; !ok {
					UnusedProject = append(UnusedProject, projects[k])
					Rename[projects[k].CompanyName] = true
				}
			}
		}
		SendNormalResponse(c, UnusedProject)
		return
	} else {
		queryList := strings.Split(query, "/")
		if len(queryList) > 3 {
			SendServerErrorResponse(c, "关联项目只有三级", nil)
			return
		}
		err := QueryList(&selector, &standardGroupList)
		if err != nil {
			SendServerErrorResponse(c, "查询标准图组失败", err)
			return
		}
		for index, filter := range queryList {
			switch index {
			case 0:
				selector["company_name"] = filter
				selector["select"] = []string{"id,name"}
			case 1:
				selector["name"] = filter
				selector["select"] = []string{"id,scan_type"}
			}
		}
		err = QueryList(&selector, &projects)
		if err != nil {
			SendServerErrorResponse(c, "查询项目列表失败", err)
			return
		}
		for i := 0; i < len(projects); i++ {
			TypeStrArr := strings.Split(projects[i].ScanType, "、")
			for _, v := range standardGroupList {
				//对已经关联完毕的数据进行匹配
				if v.ProjectID == projects[i].ID {
					//对已经关联的去除
					for i := 0; i < len(TypeStrArr); i++ {
						if TypeStrArr[i] == v.ScanType {
							TypeStrArr = append(TypeStrArr[:i], TypeStrArr[i+1:]...)
						}
					}
					//防止数组只有一个元素从而去除失败
					if len(TypeStrArr) == 1 && TypeStrArr[0] == v.ScanType {
						TypeStrArr = []string{}
					}
				}
			}
			projects[i].ScanType = strings.Join(TypeStrArr, "、")
			if projects[i].ScanType != "" {
				UnusedProject = append(UnusedProject, projects[i])
			}
		}
		if UnusedProject == nil {
			//只传递companyName的话scanType为空，只返回具体项目
			UnusedProject = projects
		}
		returnInfo := make([]ProjectFront, 0, len(projects))
		for i, project := range UnusedProject {
			scanTypeConvert(&projects[i].ScanType)
			scanTypeList := strings.Split(projects[i].ScanType, "、")
			for _, scanType := range scanTypeList {
				project.ScanType = scanType
				projectFront := ProjectFront{
					Project: project,
					ScanID:  fmt.Sprintf("%d_%s", project.ID, scanType),
				}
				if len(queryList) == 2 {
					projectFront.ID = 0
				}
				returnInfo = append(returnInfo, projectFront)
			}
		}
		SendNormalResponse(c, &returnInfo)
	}
}
