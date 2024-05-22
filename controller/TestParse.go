package controller

//
//import (
//	. "demo-go/model"
//	. "demo-go/utils"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"strconv"
//	"strings"
//)
//
//func TestParseController(c *gin.Context) {
//	var requestBody = &NewParseRequest{}
//	var err error
//	if err := c.BindJSON(&requestBody); err != nil {
//		SendParameterResponse(c, "传入参数错误", err)
//	}
//	//基本信息
//	if requestBody.ParseName == "" {
//		SendParameterResponse(c, "解析方案名称不可为空", nil)
//		return
//	}
//	if requestBody.ProjectID == 0 {
//		SendParameterResponse(c, "关联项目ID不可为空", nil)
//		return
//	}
//	//图片ID解析
//	if requestBody.ImageID == "" {
//		SendParameterResponse(c, "解析ID图片不可为空", nil)
//		return
//	}
//	if requestBody.IDConnectSymbol == "" {
//		SendParameterResponse(c, "ID连接符号不可为空", nil)
//		return
//	}
//	if requestBody.IDConnectSymbol == "other" {
//		if requestBody.OtherConnectSymbol == "" {
//			SendParameterResponse(c, "其它连接符号不可为空", nil)
//			return
//		}
//	}
//	// 文件夹解析
//	if requestBody.FileDir == "" {
//		SendParameterResponse(c, "目录名字不可为空", nil)
//		return
//	}
//	if requestBody.IDConnectSymbolForDir == "" {
//		SendParameterResponse(c, "连接符号不可为空", nil)
//		return
//	}
//	if requestBody.IDConnectSymbolForDir == "other" {
//		if requestBody.OtherConnectSymbolForDir == "" {
//			SendParameterResponse(c, "其它连接符号不可为空", nil)
//			return
//		}
//	}
//	//判断关联项目是否存在
//	var project Project
//	if err := QueryEntity(requestBody.ProjectID, &project); err != nil {
//		SendServerErrorResponse(c, "查找关联项目失败", nil)
//		return
//	}
//
//	//判断解析方案是否重名
//	var samePares Parse
//	filter := make(map[string]interface{})
//	filter["parse_name"] = requestBody.ParseName
//	if err := QueryEntityByFilter(&filter, samePares); err != nil {
//		SendServerErrorResponse(c, "查找解析方案失败", err)
//		return
//	}
//	if samePares.ID > 0 {
//		if requestBody.ID == 0 {
//			SendServerErrorResponse(c, "查找解析方案失败", nil)
//			return
//		} else if requestBody.ID != samePares.ID {
//			//用来更新用
//			SendServerErrorResponse(c, "查找解析方案失败", nil)
//			return
//		}
//	}
//
//	//解析图片名称
//	//解析图片名称
//	imageID := requestBody.ImageID                       //图片ID
//	idConnectSymbol := requestBody.IDConnectSymbol       //ID连接符号
//	otherConnectSymbol := requestBody.OtherConnectSymbol //其它连接符号
//	continueSymbol := requestBody.ContinueSymbol
//	delimiter := idConnectSymbol
//	if idConnectSymbol == "other" {
//		delimiter = otherConnectSymbol
//	}
//	if continueSymbol {
//		delimiter = strings.Split(delimiter, "")[0]
//	}
//	imageIDSep := imageID
//	imageSuffix := "B.png"
//	for _, imageType := range []string{
//		"B.png", "B.bmp", "B.jpg", "B.jpeg", "D.tif", "D.dat",
//	} {
//		if strings.HasSuffix(imageID, imageType) {
//			imageIDSep = strings.Split(imageID, delimiter+imageType)[0]
//			imageSuffix = imageType
//			break
//		}
//	}
//	imageIDSplit := strings.Split(imageIDSep, delimiter)
//	var indexList []string
//	for key := range imageIDSplit {
//		indexList = append(indexList, strconv.Itoa(key+1))
//	}
//	imageIDParse := strings.Join(indexList, delimiter)
//	//判断解析方案分隔符是否合理
//	if imageIDSplit[0] == imageIDSep {
//		SendServerErrorResponse(c, "请输入有效的分隔符", nil)
//		return
//	}
//	//解析文件夹名称
//	filePathSplit := strings.Split(requestBody.FileDir, "/") //路径
//	fileDir := filePathSplit[len(filePathSplit)-1]           //文件名
//	idConnectSymbolForDir := requestBody.IDConnectSymbolForDir
//	otherConnectSymbolForDir := requestBody.OtherConnectSymbolForDir
//	continueSymbolForDir := requestBody.ContinueSymbolForDir
//	delimiterForDir := idConnectSymbolForDir
//	if idConnectSymbolForDir == "other" {
//		delimiterForDir = otherConnectSymbolForDir
//	}
//	if continueSymbolForDir {
//		delimiterForDir = strings.Split(delimiterForDir, "")[0]
//	}
//
//	fileDirSplit := strings.Split(fileDir, delimiterForDir)
//	var indexListForDir []string
//	for key := range fileDirSplit {
//		indexListForDir = append(indexListForDir, strconv.Itoa(key+1))
//	}
//	fileDirParse := strings.Join(indexListForDir, delimiterForDir)
//	//判断解析方案分隔符是否合理
//	if fileDirSplit[0] == fileDir {
//		SendServerErrorResponse(c, "请输入有效的分隔符", nil)
//		return
//	}
//
//	if requestBody.ID == 0 {
//		transaction := DB.Begin()
//		//新增图片ID解析方案
//		var newParse Parse
//		newParse.ParseName = requestBody.ParseName
//		newParse.ProjectID = requestBody.ProjectID
//		newParse.Comment = requestBody.Comment
//		newParse.ProjectInfo = project.Info()
//		newParse.ImageID = imageIDSep
//		newParse.ImageIDParse = imageIDParse
//		newParse.IDConnectSymbol = requestBody.IDConnectSymbol
//		newParse.OtherConnectSymbol = requestBody.OtherConnectSymbol
//		newParse.ContinueSymbol = requestBody.ContinueSymbol
//		newParse.ImageSuffix = imageSuffix
//		newParse.IDIndex = uint(len(imageIDSplit))
//		newParse.FileDir = fileDir
//		newParse.FileDirParse = fileDirParse
//		newParse.IDConnectSymbolForDir = requestBody.IDConnectSymbolForDir
//		newParse.OtherConnectSymbolForDir = requestBody.OtherConnectSymbolForDir
//		newParse.ContinueSymbolForDir = requestBody.ContinueSymbolForDir
//		newParse.IDIndexForDir = uint(len(fileDirSplit))
//		err = CreateEntity(transaction, &newParse)
//		if err != nil {
//			transaction.Rollback()
//			SendServerErrorResponse(c, "创建解析方案-写入数据库失败", err)
//			return
//		}
//		operationLog := "新增解析方案：解析方案名称:" + newParse.ParseName + "; 解析方案对应的项目ID：" + fmt.Sprintf("%s", newParse.ProjectID) + "; "
//		err = Log(c, "新增解析方案", "标准化管理", operationLog, 2)
//		if err != nil {
//			SendServerErrorResponse(c, "", err)
//			return
//		}
//		//新增图片结构关联方案
//		parseImageID, err := GetOrCreateImageIDDecode(transaction, newParse, "")
//		if err != nil {
//			transaction.Rollback()
//			SendServerErrorResponse(c, "新增图片结构关联方案失败", err)
//			return
//		}
//
//		//新增文件夹结构关联方案
//		parseFileDir, err := GetOrCreateFileDirDecode(transaction, newParse, "")
//		if err != nil {
//			transaction.Rollback()
//			SendServerErrorResponse(c, "新增文件夹结构关联方案失败", err)
//			return
//		}
//		transaction.Commit()
//
//		var parseDetails ParseDetails
//		parseDetails.Parse = newParse
//		parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
//		parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
//		SendNormalResponse(c, parseDetails)
//		return
//	} else {
//		//更新解析方案
//		var parseDetails ParseDetails
//		var parse Parse
//
//		err = QueryEntity(requestBody.ID, &parse)
//		if err != nil {
//			SendServerErrorResponse(c, "读取解析方案失败", err)
//			return
//		}
//
//		transaction := DB.Begin()
//		operationLog := "更新解析方案：解析方案名称:" + parse.ParseName + "; 解析方案对应的项目ID：" + fmt.Sprintf("%s", parse.ProjectID) + "; "
//		err = Log(c, "更新解析方案", "解析方案库", operationLog, 2, transaction)
//		if err != nil {
//			transaction.Rollback()
//			SendServerErrorResponse(c, "", err)
//			return
//		}
//		selector := make(map[string]interface{})
//		selector["parse_id"] = parse.ID
//		fields := make(map[string]interface{})
//		option := requestBody.Option
//		if option == "basic" {
//			parse.ParseName = requestBody.ParseName
//			parse.ProjectID = requestBody.ProjectID
//			parse.Comment = requestBody.Comment
//			parse.ProjectInfo = project.Info()
//			err = UpdateEntity(transaction, &parse)
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新解析方案失败", err)
//				return
//			}
//			parseDetails.Parse = parse
//			parseImageID, err := GetOrCreateImageIDDecode(transaction, parse, "get")
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
//				return
//			}
//			parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
//			parseFileDir, err := GetOrCreateFileDirDecode(transaction, parse, "get")
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新文件夹结构关联方案失败", err)
//				return
//			}
//			parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
//
//			//更新与解析方案关联的字典表、匹配规则
//			fields["parse_info"] = parse.ParseName
//			err = UpdateFields(transaction, &MatchRule{}, &selector, &fields)
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新匹配规则失败", err)
//				return
//			}
//			err = UpdateFields(transaction, &Dict{}, &selector, &fields)
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新字典表失败", err)
//				return
//			}
//		} else if option == "image" {
//			parse.ImageID = imageIDSep
//			parse.ImageIDParse = imageIDParse
//			parse.IDIndex = uint(len(imageIDSplit))
//			parse.IDConnectSymbol = requestBody.IDConnectSymbol
//			parse.OtherConnectSymbol = requestBody.OtherConnectSymbol
//			err = UpdateEntity(transaction, &parse)
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新解析方案失败", err)
//				return
//			}
//
//			if parse.ContinueSymbol != requestBody.ContinueSymbol {
//				err := UpdateFields(transaction, &Parse{}, &map[string]interface{}{"ID": parse.ID}, &map[string]interface{}{"continue_symbol": requestBody.ContinueSymbol})
//				if err != nil {
//					transaction.Rollback()
//					SendServerErrorResponse(c, "更新解析方案失败", err)
//					return
//				}
//			}
//
//			parseDetails.Parse = parse
//
//			//更新图片结构关联方案
//			parseImageID, err := GetOrCreateImageIDDecode(transaction, parse, "update")
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
//				return
//			}
//
//			parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
//			parseFileDir, err := GetOrCreateFileDirDecode(transaction, parse, "get")
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
//				return
//			}
//			parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
//		} else if option == "dir" {
//			parse.FileDir = fileDir
//			parse.FileDirParse = fileDirParse
//			parse.IDIndexForDir = uint(len(fileDirSplit))
//			parse.IDConnectSymbolForDir = requestBody.IDConnectSymbolForDir
//			parse.OtherConnectSymbolForDir = requestBody.OtherConnectSymbolForDir
//			err = UpdateEntity(transaction, &parse)
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新解析方案失败", err)
//				return
//			}
//			if parse.ContinueSymbolForDir != requestBody.ContinueSymbolForDir {
//				err := UpdateFields(transaction, &Parse{}, &map[string]interface{}{"ID": parse.ID}, &map[string]interface{}{"continue_symbol": requestBody.ContinueSymbol})
//				if err != nil {
//					transaction.Rollback()
//					SendServerErrorResponse(c, "更新解析方案失败", err)
//					return
//				}
//			}
//			parseDetails.Parse = parse
//			parseImageID, err := GetOrCreateImageIDDecode(transaction, parse, "get")
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
//				return
//			}
//			parseDetails.ImageIDParseResult = parseImageID.ImageIDParseResult
//
//			更新目录结构关联方案
//			mode := "update"
//			parseFileDir, err := GetOrCreateFileDirDecode(transaction, parse, mode)
//			if err != nil {
//				transaction.Rollback()
//				SendServerErrorResponse(c, "更新图片结构关联方案失败", err)
//				return
//			}
//
//			parseDetails.FileDirParseResult = parseFileDir.FileDirParseResult
//		} else {
//			SendServerErrorResponse(c, fmt.Sprintf("不能识别的option:%s", option), nil)
//		}
//
//		transaction.Commit()
//
//		SendNormalResponse(c, parseDetails)
//		return
//	}
//}
//}
