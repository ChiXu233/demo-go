package controller

import (
	. "demo-go/model"
	. "demo-go/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wonderivan/logger"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

func CreateOrUpdateParseController(c *gin.Context) {
	var requestBody = &NewParseRequest{}
	var err error
	err = c.BindJSON(&requestBody)
	if err != nil {
		SendServerErrorResponse(c, "参数解析失败", err)
		return
	}

	//判断基本信息
	if requestBody.ParseName == "" {
		SendParameterResponse(c, "解析方案名称不能为空", err)
		return
	}
	if requestBody.ProjectID == 0 {
		SendParameterResponse(c, "解析方案关联projectID不能为空", err)
		return
	}
	if requestBody.ImageID == "" {
		SendParameterResponse(c, "解析图片ID不能为空", err)
		return
	}
	if requestBody.IDConnectSymbol == "" {
		SendParameterResponse(c, "ID连接符号不能为空", err)
		return
	}
	if requestBody.IDConnectSymbol == "other" {
		if requestBody.OtherConnectSymbol == "" {
			SendParameterResponse(c, "其他连接符号不可为空", err)
			return
		}
	}

	//文件夹解析
	if requestBody.FileDir == "" {
		SendParameterResponse(c, "目录名不能为空", err)
		return
	}
	if requestBody.IDConnectSymbolForDir == "" {
		SendParameterResponse(c, "目录连接符号不能为空", err)
		return
	}
	if requestBody.IDConnectSymbolForDir == "other" {
		if requestBody.OtherConnectSymbolForDir == "" {
			SendParameterResponse(c, "其他连接符号不可为空", err)
			return
		}
	}

	//判断关联项目是否存在
	var project Project
	err = QueryEntity(requestBody.ProjectID, &project)
	if err != nil {
		SendServerErrorResponse(c, "项目查找失败", err)
		return
	}
	//判断同名解析方案是否存在
	var sameParse Parse
	selector := make(map[string]interface{})
	selector["parse_name"] = requestBody.ParseName
	err = QueryEntityByFilter(&selector, &sameParse)
	if err != nil {
		SendServerErrorResponse(c, "查找解析方案失败", err)
		return
	}
	if sameParse.ID > 0 {
		if requestBody.ID == 0 {
			SendServerErrorResponse(c, "解析方案已存在", err)
			return
		} else if sameParse.ID != requestBody.ID {
			SendServerErrorResponse(c, "解析方案已存在", err)
			return
		}
	}

	//解析图片名称
	imageID := requestBody.ImageID
	idConnectSymbol := requestBody.IDConnectSymbol
	otherConnectSymbol := requestBody.OtherConnectSymbol
	continueSymbol := requestBody.ContinueSymbol
	delimiter := idConnectSymbol
	if idConnectSymbol == "other" {
		delimiter = otherConnectSymbol
	}
	if continueSymbol {
		delimiter = strings.Split(delimiter, "")[0]
	}
	imageIDSep := imageID
	imageSuffix := "B.png"
	for _, imageType := range []string{
		"B.png", "B.bmp", "B.jpg", "B.jpeg", "D.tif", "D.dat",
	} {
		if strings.HasSuffix(imageID, imageType) {
			imageIDSep = strings.Split(imageID, delimiter+imageType)[0]
			imageSuffix = imageType
			break
		}
	}
	imageIDSplit := strings.Split(imageIDSep, delimiter)
	var indexList []string
	for key := range imageIDSplit {
		indexList = append(indexList, strconv.Itoa(key+1))
	}
	imageIDParse := strings.Join(indexList, delimiter)
	//判断解析方案分隔符是否合理
	if imageIDSplit[0] == imageIDSep {
		SendServerErrorResponse(c, "请输入有效分割符", err)
		return
	}

	//解析文件夹名称
	filePathSplit := strings.Split(requestBody.FileDir, "/")
	fileDir := filePathSplit[len(filePathSplit)-1]
	idConnectSymbolForDir := requestBody.IDConnectSymbolForDir
	otherConnectSymbolForDir := requestBody.OtherConnectSymbolForDir
	continueSymbolForDir := requestBody.ContinueSymbolForDir
	delimiterForDir := idConnectSymbolForDir
	if idConnectSymbolForDir == "other" {
		delimiterForDir = otherConnectSymbolForDir
	}
	if continueSymbolForDir {
		delimiterForDir = strings.Split(delimiterForDir, "")[0]
	}

	fileDirSplit := strings.Split(fileDir, delimiterForDir)
	var indexListForDir []string
	for key := range fileDirSplit {
		indexListForDir = append(indexListForDir, strconv.Itoa(key+1))
	}
	fileDirParse := strings.Join(indexListForDir, delimiterForDir)
	//判断解析方案分隔符是否合理
	if fileDirSplit[0] == fileDir {
		SendServerErrorResponse(c, "请输入有效的分隔符", err)
		return
	}

	if requestBody.ID == 0 {
		transaction := DB.Begin()
		//新增图片ID解析方案
		var newParse Parse
		newParse.ParseName = requestBody.ParseName
		newParse.ProjectID = requestBody.ProjectID
		newParse.Comment = requestBody.Comment
		newParse.ProjectInfo = project.Info()
		newParse.ImageID = imageIDSep
		newParse.ImageIDParse = imageIDParse
		newParse.IDConnectSymbol = requestBody.IDConnectSymbol
		newParse.OtherConnectSymbol = requestBody.OtherConnectSymbol
		newParse.ContinueSymbol = requestBody.ContinueSymbol
		newParse.ImageSuffix = imageSuffix
		newParse.IDIndex = uint(len(imageIDSplit))
		newParse.FileDir = fileDir
		newParse.FileDirParse = fileDirParse
		newParse.IDConnectSymbolForDir = requestBody.IDConnectSymbolForDir
		newParse.OtherConnectSymbolForDir = requestBody.OtherConnectSymbolForDir
		newParse.ContinueSymbolForDir = requestBody.ContinueSymbolForDir
		newParse.IDIndexForDir = uint(len(fileDirSplit))
		err = CreateEntity(transaction, &newParse)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "创建解析方案-写入数据库失败", err)
			return
		}
		operationLog := "新增解析方案:解析方案名称：" + newParse.ParseName + "; 解析方案对应的项目ID：" + String(newParse.ProjectID) + "；"
		err = Log(c, "新增解析方案", "标准化管理", operationLog, 2)
		if err != nil {
			SendServerErrorResponse(c, "", err)
			return
		}
		//新增图片结构关联方案
		parseImageID, err := GetOrCreateImageIDDecode(transaction, newParse, "")
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "新增图片结构关联方案失败", err)
			return
		}

		//新增文件夹结构关联方案
		parseFileDir, err := GetOrCreateFileDirDecode(transaction, newParse, "")
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "新增文件夹结构关联方案失败", err)
			return
		}
		transaction.Commit()

		var parseDetails ParseDetails
		parseDetails.Parse = newParse
		parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
		parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
		SendNormalResponse(c, parseDetails)
		return
	} else {
		//更新解析方案
		var parseDetails ParseDetails
		var parse Parse

		err = QueryEntity(requestBody.ID, &parse)
		if err != nil {
			SendServerErrorResponse(c, "读取解析方案失败", err)
			return
		}

		transaction := DB.Begin()
		operationLog := "更新解析方案：解析方案名称:" + parse.ParseName + "; 解析方案对应的项目ID：" + String(parse.ProjectID) + "; "
		err = Log(c, "更新解析方案", "解析方案库", operationLog, 2, transaction)
		if err != nil {
			transaction.Rollback()
			SendServerErrorResponse(c, "", err)
			return
		}
		selector := make(map[string]interface{})
		selector["parse_id"] = parse.ID
		fields := make(map[string]interface{})
		option := requestBody.Option
		if option == "basic" {
			parse.ParseName = requestBody.ParseName
			parse.ProjectID = requestBody.ProjectID
			parse.Comment = requestBody.Comment
			parse.ProjectInfo = project.Info()
			err = UpdateEntity(transaction, &parse)
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新解析方案失败", err)
				return
			}
			parseDetails.Parse = parse
			parseImageID, err := GetOrCreateImageIDDecode(transaction, parse, "get")
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
				return
			}
			parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
			parseFileDir, err := GetOrCreateFileDirDecode(transaction, parse, "get")
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新文件夹结构关联方案失败", err)
				return
			}
			parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult

			//更新与解析方案关联的字典表、匹配规则
			fields["parse_info"] = parse.ParseName
			err = UpdateFields(transaction, &MatchRule{}, &selector, &fields)
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新匹配规则失败", err)
				return
			}
			err = UpdateFields(transaction, &Dict{}, &selector, &fields)
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新字典表失败", err)
				return
			}
		} else if option == "image" {
			parse.ImageID = imageIDSep
			parse.ImageIDParse = imageIDParse
			parse.IDIndex = uint(len(imageIDSplit))
			parse.IDConnectSymbol = requestBody.IDConnectSymbol
			parse.OtherConnectSymbol = requestBody.OtherConnectSymbol
			err = UpdateEntity(transaction, &parse)
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新解析方案失败", err)
				return
			}

			if parse.ContinueSymbol != requestBody.ContinueSymbol {
				err := UpdateFields(transaction, &Parse{}, &map[string]interface{}{"ID": parse.ID}, &map[string]interface{}{"continue_symbol": requestBody.ContinueSymbol})
				if err != nil {
					transaction.Rollback()
					SendServerErrorResponse(c, "更新解析方案失败", err)
					return
				}
			}

			parseDetails.Parse = parse

			//更新图片结构关联方案
			parseImageID, err := GetOrCreateImageIDDecode(transaction, parse, "update")
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
				return
			}

			parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
			parseFileDir, err := GetOrCreateFileDirDecode(transaction, parse, "get")
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
				return
			}
			parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
		} else if option == "dir" {
			parse.FileDir = fileDir
			parse.FileDirParse = fileDirParse
			parse.IDIndexForDir = uint(len(fileDirSplit))
			parse.IDConnectSymbolForDir = requestBody.IDConnectSymbolForDir
			parse.OtherConnectSymbolForDir = requestBody.OtherConnectSymbolForDir
			err = UpdateEntity(transaction, &parse)
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新解析方案失败", err)
				return
			}
			if parse.ContinueSymbolForDir != requestBody.ContinueSymbolForDir {
				err := UpdateFields(transaction, &Parse{}, &map[string]interface{}{"ID": parse.ID}, &map[string]interface{}{"continue_symbol": requestBody.ContinueSymbol})
				if err != nil {
					transaction.Rollback()
					SendServerErrorResponse(c, "更新解析方案失败", err)
					return
				}
			}
			parseDetails.Parse = parse
			parseImageID, err := GetOrCreateImageIDDecode(transaction, parse, "get")
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
				return
			}
			parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult

			//更新目录结构关联方案
			mode := "update"
			parseFileDir, err := GetOrCreateFileDirDecode(transaction, parse, mode)
			if err != nil {
				transaction.Rollback()
				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
				return
			}

			parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
		} else {
			SendServerErrorResponse(c, fmt.Sprintf("不能识别的option:%s", option), nil)
		}

		transaction.Commit()

		SendNormalResponse(c, parseDetails)
		return
	}

}

func GetOrCreateImageIDDecode(transaction *gorm.DB, parse Parse, mode string) (ParseImageID, error) {
	var parseImageID ParseImageID
	imageID := parse.ImageID
	var parseDecodeList []ParseDecode
	if mode == "get" {
		selector := make(map[string]interface{})
		selector["parse_id"] = parse.ID
		selector["order"] = "parse_code"
		selector["field_source"] = 1
		err := QueryList(&selector, &parseDecodeList)
		if err != nil {
			logger.Error("查询图片结构解析方案失败", err)
			return parseImageID, err
		}
		for _, decode := range parseDecodeList {
			parseImageID.ImageIDParseResult.SplitResult = append(parseImageID.ImageIDParseResult.SplitResult, decode.ImageField)
			parseImageID.ImageIDParseResult.ParseCode = append(parseImageID.ImageIDParseResult.ParseCode, decode.ParseCode)
			parseImageID.ImageIDParseResult.ParseName = append(parseImageID.ImageIDParseResult.ParseName, decode.ParseName)
		}
		return parseImageID, nil
	} else {
		var separator string
		if parse.IDConnectSymbol == "other" {
			separator = parse.OtherConnectSymbol
		} else {
			separator = parse.IDConnectSymbol
		}
		imageIDSplit := strings.Split(imageID, separator)
		for i := 0; i < len(imageIDSplit); i++ {
			var parseDecode ParseDecode
			parseDecode.ProjectID = parse.ProjectID
			parseDecode.ParseID = parse.ID
			parseDecode.FieldSource = 1
			parseDecode.ImageField = imageIDSplit[i]
			parseDecode.ParseCode = uint(i + 1)
			parseDecode.Delimiter = separator
			parseDecodeList = append(parseDecodeList, parseDecode)
			parseImageID.ImageIDParseResult.SplitResult = append(parseImageID.ImageIDParseResult.SplitResult, imageIDSplit[i])
			parseImageID.ImageIDParseResult.ParseCode = append(parseImageID.ImageIDParseResult.ParseCode, uint(i+1))
			parseImageID.ImageIDParseResult.ParseName = append(parseImageID.ImageIDParseResult.ParseName, "")
		}
		if mode == "update" {
			filter := make(map[string]interface{})
			filter["parse_id"] = parse.ID
			filter["field_source"] = 1
			_, err := DeleteEntities(transaction, &filter, &ParseDecode{})
			if err != nil {
				logger.Error("删除旧的图片结构解析方案失败", err)
				return parseImageID, err
			}
		}
		err := CreateEntities(transaction, &parseDecodeList)
		if err != nil {
			logger.Error("新增图片结构关联方案失败", err)
			return parseImageID, err
		}
		return parseImageID, nil
	}

}

func GetOrCreateFileDirDecode(transaction *gorm.DB, parse Parse, mode string) (ParseFileDir, error) {
	fileDir := parse.FileDir
	var parseFileDir ParseFileDir
	var parseDecodeList []ParseDecode
	if mode == "get" {
		selector := make(map[string]interface{})
		selector["parse_id"] = parse.ID
		selector["order"] = "parse_code"
		selector["field_source"] = 2
		err := QueryList(&selector, &parseDecodeList)
		if err != nil {
			logger.Error("查询图片结构解析方案失败", err)
			return parseFileDir, err
		}
		for _, decode := range parseDecodeList {
			parseFileDir.FileDirParseResult.SplitResult = append(parseFileDir.FileDirParseResult.SplitResult, decode.ImageField)
			parseFileDir.FileDirParseResult.ParseCode = append(parseFileDir.FileDirParseResult.ParseCode, decode.ParseCode)
			parseFileDir.FileDirParseResult.ParseName = append(parseFileDir.FileDirParseResult.ParseName, decode.ParseName)
			parseFileDir.FileDirParseResult.NeedEnter = append(parseFileDir.FileDirParseResult.NeedEnter, decode.NeedEnter)
		}
		return parseFileDir, nil
	} else {
		var separator string
		if parse.IDConnectSymbolForDir == "other" {
			separator = parse.OtherConnectSymbolForDir
		} else {
			separator = parse.IDConnectSymbolForDir
		}
		parseFileDir.FileDir = parse.FileDir
		parseFileDir.FileDirParse = parse.FileDirParse
		fileDirSplit := strings.Split(fileDir, separator)

		for i := 0; i < len(fileDirSplit); i++ {
			var parseDecode ParseDecode
			parseDecode.ProjectID = parse.ProjectID
			parseDecode.ParseID = parse.ID
			parseDecode.FieldSource = 2
			parseDecode.ImageField = fileDirSplit[i]
			parseDecode.ParseCode = uint(i + 1)
			parseDecode.Delimiter = separator
			parseDecode.NeedEnter = true
			parseDecodeList = append(parseDecodeList, parseDecode)
			parseFileDir.FileDirParseResult.SplitResult = append(parseFileDir.FileDirParseResult.SplitResult, fileDirSplit[i])
			parseFileDir.FileDirParseResult.ParseCode = append(parseFileDir.FileDirParseResult.ParseCode, uint(i+1))
			parseFileDir.FileDirParseResult.ParseName = append(parseFileDir.FileDirParseResult.ParseName, "")
			parseFileDir.FileDirParseResult.NeedEnter = append(parseFileDir.FileDirParseResult.NeedEnter, true)
		}
		if mode == "update" {
			filter := make(map[string]interface{})
			filter["parse_id"] = parse.ID
			filter["field_source"] = 2
			_, err := DeleteEntities(DB, &filter, &ParseDecode{})
			if err != nil {
				logger.Error("删除旧的图片结构解析方案失败", err)
				return parseFileDir, err
			}
		}
		err := CreateEntities(transaction, &parseDecodeList)
		if err != nil {
			logger.Error("新增图片结构关联方案失败", err)
			return parseFileDir, err
		}
		return parseFileDir, nil
	}

}
