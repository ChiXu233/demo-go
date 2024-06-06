package controller

import (
	"demo-go/config"
	. "demo-go/model"
	. "demo-go/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"gorm.io/gorm"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func CreateOrUpdateStandardGroupController(c *gin.Context) {
	var err error
	var frontStandardGroupStructure FrontStandardGroupStructure
	var standardGroup StandardGroup
	var project Project
	var standardCount int64
	err = c.BindJSON(&frontStandardGroupStructure)
	if err != nil {
		SendParameterResponse(c, "解析json数据失败", err)
		return
	}
	scanTypeConvert(&frontStandardGroupStructure.ScanType)
	standardGroup.ScanType = frontStandardGroupStructure.ScanType
	standardGroup.Name = frontStandardGroupStructure.Name
	standardGroup.Comment = frontStandardGroupStructure.Comment
	standardGroup.ProjectID = frontStandardGroupStructure.ProjectID
	standardGroup.PreprocessID = frontStandardGroupStructure.PreProcessID

	//查询是否重复
	if standardGroup.Name == "" {
		SendParameterResponse(c, "标准图组名称不能为空", nil)
		return
	}
	selector := make(map[string]interface{})
	selector["project_id"] = standardGroup.ProjectID
	if frontStandardGroupStructure.ScanType != "" {
		selector["scan_type"] = frontStandardGroupStructure.ScanType
	}
	var standardGroupList []StandardGroup
	if err = QueryList(&selector, &standardGroup); err != nil {
		SendServerErrorResponse(c, "查找项目失败", err)
		return
	}
	if len(standardGroupList) > 0 {
		SendParameterResponse(c, "此项目已经创建标准图组", err)
		return
	}

	//查询项目
	err = QueryEntity(standardGroup.ProjectID, &project)
	if err != nil {
		SendServerErrorResponse(c, "查找项目失败", err)
		return
	}
	standardGroup.ProjectInfo = project.Info()

	//查找标准图数量
	if err = DB.Model(&StandardInfo{}).Where("project_id = ?", project.ID).Count(&standardCount).Error; err != nil {
		SendServerErrorResponse(c, "查询标准图数量失败", err)
		return
	}
	number := int(standardCount)
	uploadStatus := "done"
	uploadProgress := 100
	if standardGroup.ScanType == "360" {
		//创建标准图组
		standardGroup.ScanType = "360"
		standardGroup.Numbers = 0
		standardGroup.UpdatedTime = time.Now()
		standardGroup.ProjectID = project.ID
		standardGroup.UploadStatus = uploadStatus
		standardGroup.UploadProgress = uploadProgress
		err = CreateEntity(DB, &standardGroup)
		if err != nil {
			SendServerErrorResponse(c, "创建标准图组失败", err)
			return
		}
		operationLog := fmt.Sprintf("项目：%s (%d) 创建标准图组：%s (%d)", project.Info(), project.ID, standardGroup.Name, standardGroup.ID)
		err = Log(c, "创建标准图组", "标准图管理", operationLog, 2)
		if err != nil {
			SendServerErrorResponse(c, "", err)
			return
		}
		SendNormalResponse(c, "")
		return

	} else if standardGroup.ScanType == "快扫" || standardGroup.ScanType == "rapid" {
		//uploadStatus = "doing"
		//uploadProgress = 0
		//standardDir := frontStandardGroupStructure.StandardRapidDir
		//SendParameterResponse(c, "快扫标准图录入暂不支持", nil)
		//return
		if number == 0 {
			SendParameterResponse(c, "所选路径不包含图片", err)
			return
		}

		//创建标准图组
		standardGroup.Numbers = number
		standardGroup.UpdatedTime = time.Now()
		standardGroup.ProjectID = project.ID
		standardGroup.UploadStatus = uploadStatus
		standardGroup.UploadProgress = uploadProgress
		err = CreateEntity(DB, &standardGroup)
		if err != nil {
			SendServerErrorResponse(c, "保存标准图信息失败", err)
			return
		}
	} else if standardGroup.ScanType == "accurate" || standardGroup.ScanType == "精扫" {
		//创建标准图组
		standardGroup.Numbers = number
		standardGroup.UpdatedTime = time.Now()
		standardGroup.ProjectID = project.ID
		standardGroup.UploadStatus = uploadStatus
		standardGroup.UploadProgress = uploadProgress
		//standardGroup.ParseID = parse.ID
		err = CreateEntity(DB, &standardGroup)
		if err != nil {
			SendServerErrorResponse(c, "保存标准图组信息失败", err)
			return
		}
		operationLog := fmt.Sprintf("项目: %s (%d) 创建了标准图组: %s (%d)", project.Info(), project.ID, standardGroup.Name, standardGroup.ID)
		err = Log(c, "创建标准图组", "标准图管理", operationLog, 2)
		if err != nil {
			SendServerErrorResponse(c, "", err)
			return
		}
	}
	SendNormalResponse(c, "")
	return

}

func ImportLocalStandardController(c *gin.Context) {

	var err error
	var files []string
	var standardCount int64
	var standardInfos []StandardInfo
	//读取输入路径
	//"/Users/dg2023/Desktop/BASE/"
	Filepath := c.Query("filepath")
	//"files/source_data/宁波五号线"
	flactpath := c.Query("flactpath")
	if flactpath == "" {
		SendParameterResponse(c, "文件映射路径为空", nil)
		return
	}
	//判断路径是否正确
	err = GetFiles(Filepath, true, &files)
	if err != nil {
		SendParameterResponse(c, "读取路径文件失败", err)
		return
	}

	//新建标准图组(standard_group)
	transaction := DB.Begin()
	var standardGroup StandardGroup
	var frontStandardGroupStructure FrontStandardGroupStructure
	var project Project
	//var parse Parse
	err = c.BindJSON(&frontStandardGroupStructure)
	if err != nil {
		SendParameterResponse(c, "解析json数据失败", err)
		return
	}
	scanTypeConvert(&frontStandardGroupStructure.ScanType)
	standardGroup.ScanType = frontStandardGroupStructure.ScanType
	standardGroup.Name = frontStandardGroupStructure.Name
	standardGroup.Comment = frontStandardGroupStructure.Comment
	standardGroup.ProjectID = frontStandardGroupStructure.ProjectID
	standardGroup.PreprocessID = frontStandardGroupStructure.PreProcessID
	//查询是否存在标准图组
	if standardGroup.Name == "" {
		SendServerErrorResponse(c, "标准图组的名字不能为空", err)
		return
	}
	selector := make(map[string]interface{})
	selector["project_id"] = standardGroup.ProjectID
	if frontStandardGroupStructure.ScanType != "" {
		selector["scan_type"] = frontStandardGroupStructure.ScanType
	}
	var standardGroupList []StandardGroup
	err = QueryList(&selector, &standardGroupList)
	if err != nil {
		SendServerErrorResponse(c, "查询标准图组失败", err)
		return
	}
	if len(standardGroupList) > 0 {
		SendParameterResponse(c, "此项目已创建标准图组", nil)
		return
	}
	//查询项目
	err = QueryEntity(standardGroup.ProjectID, &project)
	if err != nil {
		SendServerErrorResponse(c, "寻找项目失败", err)
		return
	}
	standardGroup.ScanType = frontStandardGroupStructure.ScanType
	standardGroup.ProjectInfo = project.Info() //查询标准图数量
	err = DB.Model(&StandardInfo{}).
		Where("project_id = ?", project.ID).Count(&standardCount).Error
	if err != nil {
		SendServerErrorResponse(c, "查询标准图数量失败", err)
		return
	}

	number := int(standardCount)
	uploadStatus := "done"
	uploadProgress := 100
	//创建标准图组
	standardGroup.Numbers = number
	standardGroup.UpdatedTime = time.Now()
	standardGroup.ProjectID = project.ID
	standardGroup.UploadStatus = uploadStatus
	standardGroup.UploadProgress = uploadProgress
	err = CreateEntity(DB, &standardGroup)
	if err != nil {
		transaction.Rollback()
		SendServerErrorResponse(c, "保存标准图组信息失败", err)
		return
	}

	//	读目录下图片列表，建(standard_info)

	standardInfos, err = GetPicFilesAndInsert(flactpath, standardGroup, &files)
	if err != nil {
		transaction.Rollback()
		SendServerErrorResponse(c, "读取文件失败", err)
		return
	}

	//	本地(.jpg+.json)，读json写（standard_item）
	err = GetJsonFilesAndInsert(standardGroup, &files, transaction, standardInfos)
	if err != nil {
		transaction.Rollback()
		SendServerErrorResponse(c, "读取json失败", err)
		return
	}
	if err = transaction.Commit().Error; err != nil {
		transaction.Rollback()
		SendServerErrorResponse(c, "事务提交失败", err)
		return
	}

	SendNormalResponse(c, standardGroup)
}

func GetPicFilesAndInsert(flactpath string, group StandardGroup, files *[]string) (info []StandardInfo, err error) {
	var filesFiltered []string
	var standardInfoList []StandardInfo
	cameraStr := ""
	index := 0
	for _, filePath := range *files {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, ".") || !strings.HasSuffix(fileName, ".jpg") {
			continue
		}
		if strings.Contains(fileName, "_compressed") {
			continue
		}
		if strings.Contains(fileName, "_compress") {
			continue
		}
		if strings.Contains(fileName, "concat") {
			continue
		}
		if strings.Contains(fileName, "concat") {
			continue
		}
		if strings.Contains(fileName, "_XRGRAY") {
			continue
		}
		filesFiltered = append(filesFiltered, filePath)
	}
	if len(filesFiltered) == 0 {
		err = errors.New("目录为空请检查")
		return nil, err
	}

	for _, v := range filesFiltered {
		index += 1
		camera := path.Base(path.Dir(v))
		if cameraStr != camera {
			//切换相机index也发生改变
			index = 1
		}
		cameraStr = camera
		outFilePath := fmt.Sprintf("%s/%s/%s", flactpath, camera, path.Base(v))
		standardInfo := StandardInfo{
			ProjectID: group.ProjectID,
			InfoModel: InfoModel{
				ImageID:          camera + "-" + strings.Split(path.Base(v), ".")[0],
				ImageURL:         fmt.Sprintf("http://%s:%d/%s", config.Conf.APP.IP, config.Conf.APP.Port, outFilePath),
				ImageURLCompress: "",
			},
			InfoModelExtend: InfoModelExtend{
				DepthURL:        "",
				DepthRenderURL:  "",
				PointCloud:      "",
				RGBURL:          "",
				Texture16bitURL: "",
				DebugTextureURL: "",
				SourceDataPath:  "",
				Comment:         "",
			},
			ParseID:             0,
			RuleID:              0,
			StandardName:        camera + "-" + strings.Split(path.Base(v), ".")[0],
			GroupNumber:         0,
			CheckingStatus:      0,
			RunningState:        0,
			AnnotateStatus:      "",
			ReferenceStatus:     "",
			DepthRenderURL:      "",
			PointCloudURL:       "",
			TrainType:           "",
			Camera:              camera,
			ScanId:              String(index),
			IsAside:             false,
			DisplayFor3D:        false,
			AnnotatedFor3D:      false,
			CommentFor3D:        "",
			ConfigStatusForBolt: "",
			ImageQualityForBolt: false,
			DisplayForBolt:      false,
			LatestBrightness:    0,
			RelatedTrainNumber:  "",
			ImageChangeStatus:   false,
			ScanType:            "精扫",
		}
		//灰度图压缩图
		compressPath, err := CompressImage(outFilePath, false, 0)
		if err != nil {
			logger.Error("压缩数据失败 %v", err)
			continue
		}
		standardInfo.ImageURLCompress = fmt.Sprintf("http://%s:%d/%s", config.Conf.APP.IP, config.Conf.APP.Port, compressPath)
		standardInfoList = append(standardInfoList, standardInfo)
	}
	transaction := DB.Begin()
	if len(standardInfoList) > 0 {
		err = CreateEntities(transaction, &standardInfoList)
		if err != nil {
			err = errors.New("创建标准图组失败")
			return nil, err
		}
	}
	transaction.Commit()
	selector := make(map[string]interface{})
	selector["select"] = []string{"id,image_id,project_id"}

	defer func() {
		err = QueryEntityByFilter(&selector, &info)
		if err != nil {
			err = errors.New("查找group_id失败")
			return
		}
	}()

	return info, nil
}

func GetJsonFilesAndInsert(group StandardGroup, files *[]string, transaction *gorm.DB, standardInfos []StandardInfo) (err error) {
	var filesFiltered []string
	var Items []Item
	index := 0
	cameraStr := ""
	selector := make(map[string]interface{})

	for _, filePath := range *files {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, ".") || !strings.HasSuffix(fileName, ".json") {
			continue
		}
		if strings.Contains(fileName, "_compressed") {
			continue
		}
		if strings.Contains(fileName, "_compress") {
			continue
		}
		if strings.Contains(fileName, "concat") {
			continue
		}
		filesFiltered = append(filesFiltered, filePath)
	}
	if len(filesFiltered) == 0 {
		err = errors.New("目录为空请检查")
		return err
	}

	var standardGroup StandardGroup
	selector = make(map[string]interface{})
	selector["project_id"] = group.ProjectID
	err = QueryEntityByFilter(&selector, &standardGroup)
	if err != nil {
		err = errors.New("查找group_id失败")
		return
	}
	for _, v := range filesFiltered {
		var data LabelMeJson
		var standard_infoId uint
		index += 1
		camareIndex := 0
		camera := path.Base(path.Dir(v))
		if cameraStr != camera {
			index = 1
		}
		cameraStr = camera
		imagId := camera + "-" + strings.Split(path.Base(v), ".")[0]

		var DBJshapes, XBJshapes, LJshapes, CXshapes, CZshapes, ZXJshapes []Shape

		//获取相同imageID的standard_infoID
		for _, k := range standardInfos {
			if k.ImageID == imagId && k.ProjectID == uint(6475) {
				standard_infoId = k.ID
			}
		}

		fileData, err := ioutil.ReadFile(v)
		if err != nil {
			err = errors.New("读取json文件失败")
			return
		}
		err = json.Unmarshal(fileData, &data)
		if err != nil {
			err = errors.New("解码json文件失败")
			return
		}

		for _, k := range data.Shapes {
			//提取大部件小部件零件、车厢车轴转向架
			if strings.Contains(k.Label, "-") && !strings.HasSuffix(k.Label, "-centre") {
				continue
			}
			//大部件
			if strings.Contains(k.Label, "#dbj") {
				DBJshapes = append(DBJshapes, k)
			}
			//小部件
			if strings.Contains(k.Label, "#xbj") {
				XBJshapes = append(XBJshapes, k)
			}
			//车厢
			if strings.Contains(k.Label, "#cx") {
				CXshapes = append(CXshapes, k)
			}
			//车轴
			if strings.Contains(k.Label, "#cz") {
				CZshapes = append(CZshapes, k)
			}
			//转向架
			if strings.Contains(k.Label, "#zxj") {
				ZXJshapes = append(ZXJshapes, k)
			}
			//零件
			if !strings.Contains(k.Label, "#") && !strings.Contains(k.Label, "-") || strings.HasSuffix(k.Label, "-centre") {
				LJshapes = append(LJshapes, k)
			}
		}
		//查找groupID，处理点位，生成standard_item
		for _, Lshapes := range LJshapes {
			var RoiArry []float64
			area := "1"
			component := "1"
			det_type := "1"
			zxj := "0"
			cz := "0"
			cx := "0"
			//初始化
			camareIndex += 1
			standardItem := Item{
				ProjectID:       group.ProjectID,
				ScanType:        "accurate",
				PointID:         standard_infoId,
				InfoID:          0,
				Enable:          1,
				Comment:         "",
				StandardGroupID: String(standardGroup.ID),
			}
			standardItem.Roi = nil
			standardItem.RoiType = Lshapes.ShapeType
			standardItem.RoiCode = imagId + "-" + String(camareIndex)
			standardItem.RoiNumber = camareIndex
			standardItem.RoiSource = 1
			standardItem.Name = ""
			standardItem.Area = ""
			standardItem.Component = ""
			standardItem.DetType = ""
			standardItem.ErrorTypes = ""

			RoiArry = handelPoint(Lshapes.Points)
			standardItem.Roi = RoiArry

			//判断零件小部件大部件
			det_type = strings.Split(Lshapes.Label, "-")[0]
			for _, xshapes := range XBJshapes {
				//小部件包含零件
				if contain(RoiArry, handelPoint(xshapes.Points)) {
					component = xshapes.Label[4:]
				}
			}
			for _, dshapes := range DBJshapes {
				//大部件包含零件
				if contain(RoiArry, handelPoint(dshapes.Points)) {
					area = dshapes.Label[4:]
				}
			}
			for _, zxjshapes := range ZXJshapes {
				if contain(RoiArry, handelPoint(zxjshapes.Points)) {
					zxj = zxjshapes.Label[4:]
				}
			}
			for _, czshapes := range CZshapes {
				if contain(RoiArry, handelPoint(czshapes.Points)) {
					cz = czshapes.Label[3:]
				}
			}
			for _, cxshapes := range CXshapes {
				if contain(RoiArry, handelPoint(cxshapes.Points)) {
					cx = cxshapes.Label[3:]
				}
			}
			standardItem.Roi = RoiArry
			standardItem.Name = area + "-" + component + "-" + det_type
			standardItem.Area = area
			standardItem.Component = component
			standardItem.DetType = det_type
			standardItem.Position = cx + "-" + cz + "-" + zxj
			Items = append(Items, standardItem)
		}
	}
	err = CreateEntities(transaction, &Items)
	if err != nil {
		err = errors.New("新增失败")
		return err
	}

	return nil
}

// 处理点位
func handelPoint(arr [][]float64) []float64 {
	//将二维数组拆分为一纬数组
	//左下右上
	res := make([]float64, 4)
	//先取最左x坐标
	if arr[0][0] < arr[1][0] {
		res[0] = arr[0][0]
		res[2] = arr[1][0]
	} else {
		res[0] = arr[1][0]
		res[2] = arr[0][0]
	}
	//取出最下y坐标
	if arr[0][1] < arr[1][1] {
		res[1] = arr[0][1]
		res[3] = arr[1][1]
	} else {
		res[1] = arr[1][1]
		res[3] = arr[0][1]
	}
	return res
}

// 判断b是否包含a(a小 b大)
func contain(a []float64, b []float64) bool {
	if b[0] <= a[0] && b[2] >= a[2] {
		if b[1] <= a[1] && b[3] >= a[3] {
			return true
		}
	}
	return false
}

func GetJsonItesm(c *gin.Context) {
	var err error
	var files []string
	var standardInfos []StandardInfo
	var filesFiltered []string
	var Items []Item
	index := 0
	cameraStr := ""
	selector := make(map[string]interface{})

	//读取输入路径
	Filepath := c.Query("filepath")
	//判断路径是否正确
	err = GetFiles(Filepath, true, &files)
	if err != nil {
		SendParameterResponse(c, "读取路径文件失败", err)
		return
	}

	selector["select"] = []string{"id,image_id,project_id"}
	err = QueryEntityByFilter(&selector, &standardInfos)
	if err != nil {
		SendServerErrorResponse(c, "查找group_id失败", err)
		return
	}
	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, ".") || !strings.HasSuffix(fileName, ".json") {
			continue
		}
		if strings.Contains(fileName, "_compressed") {
			continue
		}
		if strings.Contains(fileName, "_compress") {
			continue
		}
		if strings.Contains(fileName, "concat") {
			continue
		}
		filesFiltered = append(filesFiltered, filePath)
	}
	if len(filesFiltered) == 0 {
		SendParameterResponse(c, "目录为空", err)
		return
	}

	var standardGroup StandardGroup
	selector = make(map[string]interface{})
	selector["project_id"] = uint(6475)
	err = QueryEntityByFilter(&selector, &standardGroup)
	if err != nil {
		SendServerErrorResponse(c, "查找project_id失败", err)
		return
	}
	for _, v := range filesFiltered {
		var data LabelMeJson
		var standard_infoId uint
		index += 1
		camareIndex := 0
		camera := path.Base(path.Dir(v))
		if cameraStr != camera {
			index = 1
		}
		cameraStr = camera
		imagId := camera + "-" + strings.Split(path.Base(v), ".")[0]

		var DBJshapes, XBJshapes, LJshapes, CXshapes, CZshapes, ZXJshapes []Shape

		//获取相同imageID的standard_infoID
		for _, k := range standardInfos {
			if k.ImageID == imagId && k.ProjectID == uint(6475) {
				standard_infoId = k.ID
			}
		}

		fileData, err := ioutil.ReadFile(v)
		if err != nil {
			err = errors.New("读取json文件失败")
			return
		}
		err = json.Unmarshal(fileData, &data)
		if err != nil {
			err = errors.New("解码json文件失败")
			return
		}

		for _, k := range data.Shapes {
			//提取大部件小部件零件、车厢车轴转向架
			if strings.Contains(k.Label, "-") && !strings.HasSuffix(k.Label, "-centre") {
				continue
			}
			//大部件
			if strings.Contains(k.Label, "#dbj") {
				DBJshapes = append(DBJshapes, k)
			}
			//小部件
			if strings.Contains(k.Label, "#xbj") {
				XBJshapes = append(XBJshapes, k)
			}
			//车厢
			if strings.Contains(k.Label, "#cx") {
				CXshapes = append(CXshapes, k)
			}
			//车轴
			if strings.Contains(k.Label, "#cz") {
				CZshapes = append(CZshapes, k)
			}
			//转向架
			if strings.Contains(k.Label, "#zxj") {
				ZXJshapes = append(ZXJshapes, k)
			}
			//零件
			if !strings.Contains(k.Label, "#") && !strings.Contains(k.Label, "-") || strings.HasSuffix(k.Label, "-centre") {
				LJshapes = append(LJshapes, k)
			}
		}
		//查找groupID，处理点位，生成standard_item
		for _, Lshapes := range LJshapes {
			var RoiArry []float64
			area := "1"
			component := "1"
			det_type := "1"
			zxj := "0"
			cz := "0"
			cx := "0"
			//初始化
			camareIndex += 1
			standardItem := Item{
				ProjectID:       uint(6475),
				ScanType:        "accurate",
				PointID:         standard_infoId,
				InfoID:          0,
				Enable:          1,
				Comment:         "",
				StandardGroupID: "2015",
			}
			standardItem.Roi = nil
			standardItem.RoiType = Lshapes.ShapeType
			standardItem.RoiCode = imagId + "-" + String(camareIndex)
			standardItem.RoiNumber = camareIndex
			standardItem.RoiSource = 1
			standardItem.Name = ""
			standardItem.Area = ""
			standardItem.Component = ""
			standardItem.DetType = ""
			standardItem.ErrorTypes = ""

			RoiArry = handelPoint(Lshapes.Points)
			standardItem.Roi = RoiArry

			//判断零件小部件大部件
			det_type = strings.Split(Lshapes.Label, "-")[0]
			for _, xshapes := range XBJshapes {
				//小部件包含零件
				if contain(RoiArry, handelPoint(xshapes.Points)) {
					component = xshapes.Label[4:]
				}
			}
			for _, dshapes := range DBJshapes {
				//大部件包含零件
				if contain(RoiArry, handelPoint(dshapes.Points)) {
					area = dshapes.Label[4:]
				}
			}
			for _, zxjshapes := range ZXJshapes {
				if contain(RoiArry, handelPoint(zxjshapes.Points)) {
					zxj = zxjshapes.Label[4:]
				}
			}
			for _, czshapes := range CZshapes {
				if contain(RoiArry, handelPoint(czshapes.Points)) {
					cz = czshapes.Label[3:]
				}
			}
			for _, cxshapes := range CXshapes {
				if contain(RoiArry, handelPoint(cxshapes.Points)) {
					cx = cxshapes.Label[3:]
				}
			}

			standardItem.Roi = RoiArry
			standardItem.Name = area + "-" + component + "-" + det_type
			standardItem.Area = area
			standardItem.Component = component
			standardItem.DetType = det_type
			standardItem.Position = cx + "-" + cz + "-" + zxj
			Items = append(Items, standardItem)
		}

	}
	transaction := DB.Begin()
	err = CreateEntities(transaction, &Items)
	if err != nil {
		transaction.Rollback()
		SendServerErrorResponse(c, "录入失败", err)
		return
	}
	transaction.Commit()
	SendNormalResponse(c, "录入成功")
	return
}

// InsertCompress 传入指定路径生成压缩图
func InsertCompress(c *gin.Context) {
	var wg sync.WaitGroup
	var err error
	var files []string
	var filesFiltered []string
	Filepath := c.Query("filepath")
	//outPath := "files/source_data/宁波五号线"
	outPath := Filepath
	if Filepath == "" {
		SendParameterResponse(c, "传入路径为空", nil)
		return
	}

	//校验路径下文件是否正确
	err = GetFiles(Filepath, true, &files)
	if err != nil {
		SendServerErrorResponse(c, "读取路径文件失败", err)
		return
	}
	//生成该路径下压缩图
	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, ".") || !strings.HasSuffix(fileName, ".jpg") {
			continue
		}
		if strings.Contains(fileName, "_compressed") {
			continue
		}
		if strings.Contains(fileName, "_compress") {
			continue
		}
		if strings.Contains(fileName, "concat") {
			continue
		}
		if strings.Contains(fileName, "concat") {
			continue
		}
		if strings.Contains(fileName, "_XRGRAY") {
			continue
		}
		filesFiltered = append(filesFiltered, filePath)
	}
	if len(filesFiltered) == 0 {
		SendParameterResponse(c, "目录为空请检查", nil)
		return
	}
	for _, v := range filesFiltered {
		outfilePath := fmt.Sprintf("/%s/%s", outPath, path.Base(v))
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = CompressImage(outfilePath, false, 0)
			if err != nil {
				fmt.Println("压缩数据失败", err)
				return
			}
		}()
		wg.Wait()
	}
	SendNormalResponse(c, "压缩成功")
}
