package model

//func ReadSourceDataFromLocalDisk(path string, input interface{}, points *[]PointModel, scanType string) error {
//	var err error
//	var files []string
//	var dataSource = "local"
//	// params 可以是*Project 也可以是location, serial的map
//	// todo default params
//	switch input.(type) {
//	case *Project:
//		inputParse := input.(*Project)
//
//		// //校验项目信息是否与本地路径信息一致
//		// //todo 这里要求固定的存储路径，比如：xxx/无锡二号线/精扫/xxx
//		deploymentFlag := false
//		deployment := AliasMapping[inputParse.Location]
//		for _, item := range strings.Split(path, "/") {
//			if strings.Contains(item, deployment) {
//				deploymentFlag = true
//				if strings.Contains(item, "metaloop") {
//					dataSource = "metaloop"
//				}
//				break
//			}
//		}
//		if !deploymentFlag {
//			return errors.New("所选路径和项目不对应, 请检查")
//		}
//
//		//精扫和快扫 分别做处理
//		switch scanType {
//		case "accurate":
//			// 每个案场在此处做不同的处理
//			var parse Parse
//			var parseDecodeList []ParseDecode
//			selector := make(map[string]interface{})
//			selector["project_id"] = inputParse.ID
//
//			// 读取匹配关系表(info), 查出此项目所有匹配关系
//
//			// 读解析方案表(parse)，获取分割符和图片后缀
//			err = QueryEntityByFilter(&selector, &parse)
//			if err != nil {
//				return err
//			}
//			if parse.ID == 0 {
//				return errors.New("can not find parsing programme")
//			}
//
//			// 然后读字典表(parse_decode), 获取车厢轴等信息所在字段
//			selector["parse_id"] = parse.ID
//			err = QueryEntityByFilter(&selector, &parseDecodeList)
//			if err != nil {
//				return err
//			}
//			// cameraIndex := -1
//			// carriageIndex := -1
//			// bogieIndex := -1
//			// axleIndex := -1
//			// positionIndex := -1
//			// for _, parseDecode := range parseDecodeList {
//			// 	index := int(parseDecode.ParseCode - 1)
//			// 	if parseDecode.ParseName == "相机" {
//			// 		cameraIndex = index
//			// 	}
//			// 	if parseDecode.ParseName == "车厢" {
//			// 		carriageIndex = index
//			// 	}
//			// 	if parseDecode.ParseName == "转向架" {
//			// 		bogieIndex = index
//			// 	}
//			// 	if parseDecode.ParseName == "轴位" {
//			// 		axleIndex = index
//			// 	}
//			// 	if parseDecode.ParseName == "左右" {
//			// 		positionIndex = index
//			// 	}
//			// }
//			idConnectSymbol := parse.IDConnectSymbol
//			imageSuffix := parse.ImageSuffix
//			depthSuffix := ""
//			rgbSuffix := ""
//			// cameraDefault := ""
//			// carriagePathIndex := 0
//			plySuffix := ""
//			idLength := 0
//			//不同案场参数微调
//			switch deployment {
//			case "无锡二号线":
//				depthSuffix = "D.tif"
//				// imageSuffix = "_" + imageSuffix
//				// 测试库里数据比较乱, 查出来可能是错的, 暂时写死吧
//				imageSuffix = "B.png"
//				idLength = 13
//			case "无锡四号线":
//				depthSuffix = "D.tif"
//				// imageSuffix = "_" + imageSuffix
//				imageSuffix = "B.png"
//				rgbSuffix = "C.bmp"
//				idLength = 7
//			case "长沙":
//				depthSuffix = ".tif"
//				// 长沙数据没有B/T的标记
//				imageSuffix = ".png"
//				idLength = 6
//			case "广州南":
//				depthSuffix = "D.tif"
//				imageSuffix = "B.png"
//				idLength = 11
//			case "北京北":
//				depthSuffix = "D.tif"
//				imageSuffix = "B.png"
//				idLength = 11
//			case "济南东":
//				depthSuffix = "D.tif"
//				imageSuffix = "B.png"
//				idLength = 11
//			case "延庆":
//				depthSuffix = "D.png"
//				imageSuffix = "B.png"
//				plySuffix = "D.ply"
//				idLength = 2
//			case "怀化":
//				depthSuffix = "D.tif"
//				imageSuffix = "B.png"
//				rgbSuffix = "C.bmp"
//				idLength = 7
//			case "徐州二号线":
//				depthSuffix = "D.tif"
//				imageSuffix = "B.png"
//				rgbSuffix = "C.bmp"
//				idLength = 7
//			case "西安":
//				depthSuffix = "D.exr"
//				imageSuffix = "R.png"
//				rgbSuffix = "R.png"
//				idLength = 1
//			}
//			//获取本地图片
//			if dataSource == "metaloop" {
//				pathList := strings.Split(path, "/")
//				datasetId, err := metaloop.B.GetDataSetVersionId(pathList[0])
//				if err != nil {
//					return err
//				}
//				err = metaloop.B.GetFilesForMetaloop(path, datasetId, true, &files)
//				if err != nil {
//					return err
//				}
//
//			} else {
//				err = GetFiles(path, true, &files)
//				if err != nil {
//					return err
//				}
//			}
//
//			fileMap := make(map[string]PointModel)
//			for _, filePath := range files {
//				fileName := filepath.Base(filePath) // "x_x_x_x_B.png"
//				splitResult := strings.Split(fileName, idConnectSymbol)
//				if idLength != 0 && idLength != len(splitResult) {
//					logger.Warn(filePath + "  id length: " + String(idLength) + " split result: " + String(len(splitResult)))
//					//logger.Error(deployment + " point name invalid: " + filePath)
//					continue
//				}
//				suffix := splitResult[len(splitResult)-1] // B.png / D.tif
//				pointName := strings.Split(fileName, idConnectSymbol+suffix)[0]
//				_, exists := fileMap[pointName]
//				if !exists {
//					// camera := cameraDefault
//					// carriage := ""
//					// bogie := ""
//					// axle := ""
//					// position := ""
//					//这里camera_index是有可能为0的
//					// if cameraIndex >= 0 {
//					// 	camera = splitResult[cameraIndex]
//					// }
//					// if carriagePathIndex != 0 {
//					// 	pathSplitResult := strings.Split(filePath, "/")
//					// 	carriage = pathSplitResult[len(pathSplitResult)+carriagePathIndex-1]
//					// } else if carriageIndex >= 0 {
//					// 	carriage = splitResult[carriageIndex]
//					// }
//					// if bogieIndex >= 0 {
//					// 	bogie = splitResult[bogieIndex]
//					// }
//					// if axleIndex >= 0 {
//					// 	axle = splitResult[axleIndex]
//					// }
//					// if positionIndex >= 0 {
//					// 	position = splitResult[positionIndex]
//					// }
//					fileMap[pointName] = PointModel{
//						PointName:       pointName,
//						ImagePath:       "",
//						DepthPath:       "",
//						RGBPath:         "",
//						IDConnectSymbol: idConnectSymbol,
//					}
//				}
//				temp := fileMap[pointName]
//				if suffix == imageSuffix {
//					temp.ImagePath = filePath
//				} else if suffix == depthSuffix {
//					temp.DepthPath = filePath
//				} else if suffix == rgbSuffix {
//					temp.RGBPath = filePath
//				} else if suffix == plySuffix {
//					temp.PointCloudPath = filePath
//				}
//				fileMap[pointName] = temp
//			}
//			for _, value := range fileMap {
//				if value.PointName != "" && value.ImagePath != "" {
//					value.DataSource = dataSource
//					*points = append(*points, value)
//				}
//			}
//			return nil
//		case "rapid":
//			//不同案场下的图片参数信息
//			imageSuffix := ".bmp"
//			scanIDIndex := 1
//			idConnectSymbol := "_"
//			//案场：无锡地铁
//			switch deployment {
//			case "无锡二号线":
//				if !strings.Contains(path, "aligned") {
//					imageSuffix = "backUp.bmp"
//				}
//			case "广州南":
//				imageSuffix = "-B.jpg"
//				idConnectSymbol = "-"
//			}
//			err = GetFiles(path, true, &files)
//			if err != nil {
//				return err
//			}
//			for _, filePath := range files {
//				// 目前方案, 对齐只生成灰度图, 导出时生成点云图, 此处只读灰度图
//				if !strings.Contains(filePath, imageSuffix) || strings.Contains(filePath, "compress") {
//					continue
//				}
//				fileName := filepath.Base(filePath) // "x_y_z.bmp" / ZXB_LC03M-001-B.jpg
//				fileNameSplit := strings.Split(fileName, idConnectSymbol)
//				pointName := strings.Split(fileName, imageSuffix)[0]
//				*points = append(*points, PointModel{
//					PointName: pointName,
//					ImagePath: filePath,
//					ScanID:    fileNameSplit[scanIDIndex],
//				})
//			}
//			return nil
//		}
//	case map[string]string:
//		return nil
//	default:
//		return nil
//	}
//	return nil
//}
