package model

import (
	. "demo-go/config"
	. "demo-go/utils"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/wonderivan/logger"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type SourceData struct {
	BaseModel
	ID        uint      `gorm:"primarykey"`
	ProjectID uint      `json:"project_id"`
	Serial    string    `json:"serial"`
	DataTime  time.Time `json:"data_time"`
	BasePath  string    `json:"base_path"`
	DataName  string    `json:"data_name"`
	DataType  string    `json:"data_type"`
	DataList  []string  `gorm:"-" json:"data_list"`
	// FaultLocation    string         `json:"fault_location"`
	Comment          string `json:"comment"`
	UploadStatus     string `json:"upload_status"`
	UploadProgress   int    `json:"upload_progress"`
	ProjectInfo      string `json:"project_info"`
	AnnotationStatus string `gorm:"default:'none'" json:"annotation_status"` // none / ongoing / done / review
	PointsNumber     int    `json:"points_number,omitempty"`
	Enable           string `gorm:"column:enable;default:'enable'" json:"-"`
	DataSource       string `gorm:"-" json:"data_source"` // 数据源类型
	ScanType         string `gorm:"scan_type" json:"scan_type"`
	GroupName        string `gorm:"column:group_name" json:"group_name"`
	CopyStatus       string `gorm:"column:copy_status;default:'none'" json:"copy_status"` // none / copied 表示是否从其他组复制进来过数据，产品需求，复制过的可以一键清除标注，其他不能
}

type NewSourceDataRequest struct {
	BaseModel
	ID        uint      `gorm:"primarykey"`
	ProjectID uint      `json:"project_id"`
	ScanType  string    `json:"scan_type"`
	Serial    string    `json:"serial"`
	DataTime  time.Time `json:"data_time"`
	BasePath  string    `json:"base_path"`
	DataName  string    `json:"data_name"`
	DataType  string    `json:"data_type"`
	DataList  []string  `gorm:"-" json:"data_list"`
	// FaultLocation    string         `json:"fault_location"`
	Comment          string `json:"comment"`
	UploadStatus     string `json:"upload_status"`
	UploadProgress   int    `json:"upload_progress"`
	ProjectInfo      string `json:"project_info"`
	AnnotationStatus string `gorm:"default:'none'" json:"annotation_status"`
	PointsNumber     int    `json:"points_number,omitempty"`
	Enable           string `gorm:"column:enable;default:'enable'" json:"-"`
	DataSource       string `gorm:"-" json:"data_source"` // 数据源类型
}

func (SourceData) TableName() string {
	return "source_data"
}

type SourceDataPoint struct {
	BaseModel
	ProjectID    uint   `gorm:"column:project_id" json:"-"`
	ScanType     string `gorm:"scan_type" json:"scan_type"`
	SourceDataID uint   `gorm:"column:source_data_id"`
	// StandardName string         `gorm:"column:standard_name" json:"-"`           // 标准图名称
	RGBURL string `gorm:"column:rgb_url" json:"rgb_url,omitempty"` // RGB图URL
	// PointCloudURL string         `gorm:"column:point_cloud_url" json:"point_cloud_url,omitempty"` // 压缩点云图URL
	PointCloud string `gorm:"column:point_cloud" json:"point_cloud"` // 点云图
	ImageURL   string `gorm:"column:image_url" json:"image_url"`     // 灰度图URL
	// CompressedImageURL 	string 	`gorm:"column:compressed_image_url" json:"compressed_image_url"` 	// 未压缩图没有URL需求，所以只存压缩图URL
	DepthURL         string `gorm:"column:depth_url" json:"depth_url,omitempty"`                 // 深度图URL
	RenderedDepthURL string `gorm:"column:rendered_depth_url" json:"depth_render_url,omitempty"` // 深度渲染图URL
	Status           string `gorm:"default:'normal'" json:"status"`                              // 是否故障数据
	DataStatus       string `gorm:"default:'normal'" json:"data_status"`                         // 数据是否正常(针对异常检测，是否可用)
	ErrorTypes       string `gorm:"column:error_types" json:"-"`                                 // 包含的故障类型
	PointModel
}

func (SourceDataPoint) TableName() string {
	return "source_data_point"
}

type SourceDataItem struct {
	BaseModel
	SourceDataID   uint            `json:"-"`
	ProjectID      uint            `json:"-"`
	PointID        uint            `json:"-"`
	FrontRoi       [][]float64     `gorm:"-" json:"roi"`
	Roi            pq.Float64Array `gorm:"column:roi;type:float8[]" json:"-"`
	RoiSource      int             `gorm:"column:roi_source;default:1" json:"roi_source"` //标记来源，1表示灰度图，2表示彩色图
	RoiType        string          `gorm:"column:roi_type" json:"roi_type"`
	Name           string          `gorm:"column:name" json:"name"`
	Area           string          `gorm:"column:area" json:"area"`
	Component      string          `gorm:"column:component" json:"component"`
	DetType        string          `gorm:"column:det_type" json:"det_type"`
	ErrorTypes     string          `gorm:"column:error_types" json:"error_types"`
	StandardItemID uint            `gorm:"column:standard_item_id" json:"standard_item_id"` // 标准图框的ID
	BreakStatus    uint            `gorm:"column:status;default:0" json:"break_status"`     // 当标准图框被删除/更新故障类型后，为异常的当前图设置状态(1表示删除，2表示更新)
	ErrorCategory  string          `gorm:"column:error_category" json:"error_category"`     // 故障类别 真实/模拟/待定

	// 需要用反射，继承后不方便
	// ItemModel
	ScanType string `gorm:"scan_type" json:"scan_type"`
}

func (SourceDataItem) TableName() string {
	return "source_data_item"
}

func (t SourceDataItem) BeforeUpdate(tx *gorm.DB) error {
	t.Name = t.Area + "-" + t.Component + "-" + t.DetType + "-" + t.ErrorTypes
	return nil
}

type SearchDict struct {
	Frontend string       `json:"frontend"`
	Backend  string       `json:"backend"`
	Contains []SearchDict `json:"contains,omitempty"`
}

// SearchDicts 对[]SearchDict排序
type SearchDicts []SearchDict

func (sd SearchDicts) Less(i, j int) bool {
	return sd[i].Frontend < sd[j].Frontend
}
func (sd SearchDicts) Len() int {
	return len(sd)
}
func (sd SearchDicts) Swap(i, j int) {
	sd[i], sd[j] = sd[j], sd[i]
}
func SortSearchDicts(searchDicts []SearchDict) {
	tmp := SearchDicts(searchDicts)
	sort.Sort(tmp)
	searchDicts = tmp
}

type MapRequestParse struct {
	Value    interface{}
	Required bool
}

type ReturnID struct {
	sourceDataID uint `gorm:"source_data_id"`
	ID           uint `gorm:"id"`
}
type SourceDataTaskQueue struct {
	SourceDataID uint `json:"source_data_id"`
}

func TaskFailed(sourceData *SourceData, err error) {
	logger.Error("%v \n %v", sourceData, err)
	sourceData.Comment += fmt.Sprintf("(System Error: %v)", err)
	sourceData.UploadStatus = "failed"
	err = UpdateEntity(DB, sourceData)
	if err != nil {
		logger.Error("E0: %v", err)
	}
	return
}

var wg sync.WaitGroup

func SourceDataTask() {
	redisKey := Conf.Redis.SourceDataTaskQueueKey
	for {
		// fmt.Println("AssessExecute start read")
		content := RedisClient.BRPop(time.Second*120, redisKey).Val()
		if content != nil {
			redisTask := SourceDataTaskQueue{}
			// why content[1] ?
			err := json.Unmarshal([]byte(content[1]), &redisTask)
			if err != nil {
				logger.Error(err)
				continue
			}
			logger.Info("DEBUGHERE consume task: %v", redisTask)
			// fmt.Println("AssessExecute get task: " + String(redisTask.TaskID))
			var project Project
			var sourceData SourceData
			sourceDataID := redisTask.SourceDataID
			if sourceDataID == 0 {
				logger.Warn("[source data task] Bad Task ID=0")
				continue
			}
			err = QueryEntity(sourceDataID, &sourceData)
			if err != nil {
				logger.Error(err)
				continue
			}
			err = QueryEntity(sourceData.ProjectID, &project)
			if err != nil {
				logger.Error(err)
				continue
			}
			scanType := sourceData.ScanType
			dataPath := path.Join(sourceData.BasePath, sourceData.DataName)
			if scanType == "accurate" {
				//err = SourceDataTaskAccurate(&sourceData, &project, dataPath)
				//if err != nil {
				//	TaskFailed(&sourceData, err)
				//}
			} else if scanType == "rapid" {
				//err = SourceDataTaskRapid(&sourceData, &project, dataPath)
				//if err != nil {
				//	TaskFailed(&sourceData, err)
				//}
			} else if scanType == "360" {
				err = SourceDataTask360(&sourceData, &project, dataPath)
				if err != nil {
					TaskFailed(&sourceData, err)
				}
			}
		}
	}
}

// TODO:实现
func SourceDataTask360(s *SourceData, p *Project, dataPath string) error {
	return nil
}

//func SourceDataTaskAccurate(sourceData *SourceData, project *Project, dataPath string) error {
//	cover := false // 是否覆盖压缩 包括灰度压缩和深度渲染
//	var points []PointModel
//	var pointsNumber int
//	progressKey := Conf.Redis.SourceDataTaskProgressKey
//	err := ReadSourceDataFromLocalDisk(dataPath, project, &points, sourceData.ScanType)
//	if err != nil {
//		logger.Error(err)
//		return err
//	}
//
//	// 遍历， 压缩， 入库
//	pointsNumber = len(points)
//	if pointsNumber == 0 {
//		return errors.New("此组数据无合法数据")
//
//	}
//	sourceData.UploadStatus = "doing"
//	sourceData.PointsNumber = pointsNumber
//	err = UpdateEntity(DB, &sourceData)
//	if err != nil {
//		return err
//	}
//	// if sourceData.UploadStatus != "doing" {
//	// 	logger.Warn("已跳过: %v", sourceData)
//	// 	return nil
//	// }
//	contentKey := "source-" + String(sourceData.ID)
//	threads := Conf.Config.WriteCount
//	pointsPerThread := pointsNumber / threads
//	// 处理进度
//	var lock sync.Mutex
//	progressMap := make(map[string]interface{})
//	progressContent := make(map[string]interface{})
//	progressContent["count"] = len(points)
//	progressContent["now"] = 0
//	progressContent["success"] = 0
//	progressContent["failed"] = 0
//	progressContentJson, err := json.Marshal(progressContent)
//	if err != nil {
//		return err
//	}
//	progressMap[contentKey] = progressContentJson
//
//	// RedisClient.HMSet(redisKey, progressMap)
//	if err = RedisClient.HMSet(progressKey, progressMap).Err(); err != nil {
//		return err
//	}
//	// 每间隔gap个，写一次redis，减少查写次数
//	gap := 1
//	for i := 0; i < threads; i++ {
//		var start, end int
//		start = i * pointsPerThread
//		if i == threads-1 {
//			end = pointsNumber
//		} else {
//			end = start + pointsPerThread
//		}
//		APP := Conf.APP
//		wg.Add(1)
//		go func(pointModelList []PointModel, sourceDataID uint, threadID int) {
//			var pointList []SourceDataPoint
//			for index, pointModel := range pointModelList {
//				// time.Sleep(1 * time.Second)
//				// logger.Info("Thread-%d index:%d", threadID, index)
//				// 先查看是否已被取消
//				commandKey := "source-" + String(sourceDataID) + "-command"
//				command := RedisClient.HMGet(progressKey, commandKey).Val()[0]
//				if command == "cancel" {
//					logger.Info("Thread-%d Source-%d 录入数据, 已取消", threadID, sourceDataID)
//					return
//				}
//				// DEBUG TEMP
//				// if index > 5 {
//				// 	break
//				// }
//				var point SourceDataPoint
//				imagePath := pointModel.ImagePath
//				// fmt.Println("point name: " + pointModel.PointName)
//				if pointModel.PointName == "" || imagePath == "" {
//					logger.Error(
//						"Thread-%d Source-%d 录入数据, 读取图片路径/名称为空, point_name: %s  image_path: %s",
//						threadID, sourceDataID, pointModel.PointName, imagePath)
//					continue
//				}
//				// 压缩图片
//				compressedPath, err := CompressImage(pointModel.ImagePath, cover, 1)
//				if err != nil {
//					logger.Error(
//						"Thread-%d Source-%d 录入数据, 压缩失败 %v", threadID, sourceDataID, err)
//					continue
//				}
//				point.ProjectID = sourceData.ProjectID
//				point.ScanType = sourceData.ScanType
//				point.SourceDataID = sourceDataID
//				point.PointModel = pointModel
//				point.ImageURL = fmt.Sprintf("http://%s:%d/%s", APP.IP, APP.Port, compressedPath)
//				if IsUrl(compressedPath) {
//					point.ImageURL = compressedPath
//				}
//				if pointModel.RGBPath != "" {
//					point.RGBURL = fmt.Sprintf("http://%s:%d/%s", APP.IP, APP.Port, pointModel.RGBPath)
//					if IsUrl(pointModel.RGBPath) {
//						point.RGBURL = pointModel.RGBPath
//					}
//					compressedPath, err := CompressImage(pointModel.RGBPath, cover, 1)
//					if err != nil {
//						logger.Error(
//							"Thread-%d Source-%d 录入数据, 压缩失败 %v", threadID, sourceDataID, err)
//						continue
//					}
//					point.RGBURL = fmt.Sprintf("http://%s:%d/%s", APP.IP, APP.Port, compressedPath)
//					if IsUrl(compressedPath) {
//						point.RGBURL = compressedPath
//					}
//				}
//				if pointModel.DepthPath != "" {
//					point.DepthURL = fmt.Sprintf("http://%s:%d/%s", APP.IP, APP.Port, pointModel.DepthPath)
//					if IsUrl(pointModel.DepthPath) {
//						point.DepthURL = pointModel.DepthPath
//					}
//					// 生成深度渲染图
//					imageURL := fmt.Sprintf("http://%s:%d/%s", APP.IP, APP.Port, pointModel.ImagePath)
//					if IsUrl(pointModel.ImagePath) {
//						imageURL = pointModel.ImagePath
//					}
//					depthRender, err := DepthRendering(point.DepthURL, imageURL, cover)
//					if err != nil {
//						logger.Error("Thread-%d Source-%d 录入数据, 深度图渲染失败%v", threadID, sourceDataID, err)
//						continue
//					}
//					point.RenderedDepthURL = depthRender
//				}
//				if pointModel.PointCloudPath != "" {
//					point.PointCloudPath = pointModel.PointCloudPath
//					point.PointCloud = "/" + pointModel.PointCloudPath
//					if IsUrl(pointModel.PointCloudPath) {
//						point.PointCloud = pointModel.PointCloudPath
//					}
//				}
//				// point.CompressedImageURL = fmt.Sprintf("http://%s:%d/%s", APP.IP, APP.Port, compressedPath)
//				// logger.Info("Thread-%d Source-%d 录入数据, 已读取 %s", threadID, sourceDataID, imagePath)
//				pointList = append(pointList, point)
//				pointNameSplitMap := make(map[string]interface{})
//				pointNameSplit := strings.Split(point.PointName, point.IDConnectSymbol)
//				for index, str := range pointNameSplit {
//					pointNameSplitMap["Field"+String(index+1)] = str
//				}
//				pointNameSplitMap["ProjectID"] = sourceData.ProjectID
//				pointNameSplitMap["SourceDataID"] = sourceDataID
//				pointNameSplitMap["PointName"] = point.PointName
//
//				// pointNameSplitMap
//				// fmt.Println(pointNameSplitMap)
//				// fmt.Println(sourceDataPointSplit)
//				increase := 1
//				if gap != 1 {
//					if (index + 1) == len(pointModelList) {
//						increase = len(pointModelList) % gap
//					} else if (index+1)%gap == 0 {
//						increase = gap
//					} else {
//						increase = 0
//					}
//				}
//				if increase != 0 {
//					lock.Lock()
//					progressContent := make(map[string]interface{})
//					progress := RedisClient.HMGet(progressKey, contentKey).Val()[0]
//					err = json.Unmarshal([]byte(progress.(string)), &progressContent)
//					if err != nil {
//						logger.Error(progress)
//						logger.Error(err)
//					}
//					now := int(progressContent["now"].(float64))
//					count := int(progressContent["count"].(float64))
//					// 更新数据库为完成状态
//					// logger.Info("%d %d", now, count)
//					if now+increase == count {
//						logger.Info("Thread-%d source_data %d 将更新数据库为完成状态", threadID, sourceDataID)
//						var querySourceData SourceData
//						err = QueryEntity(sourceData.ID, &querySourceData)
//						if err != nil {
//							logger.Error(err)
//							lock.Unlock()
//							// 可能是因为数据组已被删除
//							return
//						}
//						querySourceData.UploadStatus = "done"
//						querySourceData.UploadProgress = 100
//						err := UpdateEntity(DB, &querySourceData)
//						if err != nil {
//							logger.Error(err)
//						}
//						logger.Info("Thread-%d source_data %d 更新状态完成", threadID, querySourceData.ID)
//					}
//					progressContent["now"] = now + increase
//					logger.Info("Thread-%d Source-%d ==> %d", threadID, sourceData.ID, now+increase)
//					// temp
//					progressContent["success"] = now + increase
//					progressContentJson, err := json.Marshal(progressContent)
//					if err != nil {
//						logger.Error(err)
//					}
//					progressMap[contentKey] = progressContentJson
//					if err = RedisClient.HMSet(progressKey, progressMap).Err(); err != nil {
//						logger.Error(err)
//					}
//					logger.Info("Thread-%d Source-%d %v", threadID, sourceData.ID, string(progressContentJson))
//					lock.Unlock()
//				}
//			}
//			err = CreateEntities(DB, &pointList)
//			if err != nil {
//				logger.Error("%s %v", "录入数据, 写入数据库失败", err)
//			}
//			wg.Done()
//		}(points[start:end], sourceData.ID, i)
//	}
//	wg.Wait()
//	return nil
//}

//func SourceDataTaskRapid(sourceData *SourceData, project *Project, dataPath string) error {
//	location := AliasMapping[project.Location]
//	progressKey := Conf.Redis.SourceDataTaskProgressKey
//	contentKey := "source-" + String(sourceData.ID)
//	// cur_root / camera / deployment / save_path / callback
//	// HTTPClient(method string, url string, data interface{}, response interface{}) (int, error)
//	requestURL := fmt.Sprintf("http://%s:%d/api/align", Conf.Algorithm.IP, Conf.Algorithm.Port)
//	alignRequest := make(map[string]interface{})
//	alignRequest["deployment"] = location
//	alignRequest["source_data_id"] = sourceData.ID
//	alignRequest["url"] = fmt.Sprintf("http://%s:%d/api/source_data/rapid/callback", Conf.APP.IP, Conf.APP.Port)
//	redisContent := map[string]interface{}{}
//	progressContent := map[string]interface{}{}
//	camerasContent := map[string]interface{}{}
//	// 快扫的进度因为需要合并计算，总total和now直接按百分制存储，具体进度在cameras内
//	progressContent["count"] = 100
//	progressContent["now"] = 0
//	switch location {
//	case "无锡二号线":
//		fs, err := ioutil.ReadDir(dataPath)
//		if err != nil {
//			return err
//		}
//		for _, f := range fs {
//			if StringIn(f.Name(), []string{"KS", "ks", "midData"}) {
//				dataPath = path.Join(dataPath, f.Name())
//				break
//			}
//		}
//		// 遍历文件，获取总数
//		var originPoints []PointModel
//		err = ReadSourceDataFromLocalDisk(dataPath, project, &originPoints, sourceData.ScanType)
//		if err != nil {
//			return err
//		}
//		savePath := path.Join(dataPath, "aligned")
//		alignRequest["cur_root"] = dataPath
//		alignRequest["save_path"] = savePath
//		alignRequest["camera"] = "WXDT"
//		var response map[string]interface{}
//		code, err := HTTPClient("POST", requestURL, alignRequest, &response)
//		if code != 200 || err != nil {
//			return err
//		}
//		camerasContent["WXDT"] = map[string]interface{}{
//			"total":      len(originPoints) * 3,
//			"aligned":    0,
//			"status":     "doing",
//			"compressed": 0,
//			"save_path":  savePath,
//		}
//		progressContent["total_number"] = len(originPoints) * 3
//		progressContent["cameras"] = camerasContent
//		progressContentJson, err := json.Marshal(progressContent)
//		if err != nil {
//			return err
//		}
//		redisContent[contentKey] = progressContentJson
//
//		// RedisClient.HMSet(redisKey, progressMap)
//		if err = RedisClient.HMSet(progressKey, redisContent).Err(); err != nil {
//			return err
//		}
//	case "北京北":
//		return errors.New("未支持北京北快扫")
//	case "广州南":
//		rd, err := ioutil.ReadDir(dataPath)
//		if err != nil {
//			return err
//		}
//		for _, n := range rd {
//			if StringIn(n.Name(), []string{"ks", "KS"}) {
//				dataPath = path.Join(dataPath, n.Name())
//			}
//		}
//		cameras, err := ioutil.ReadDir(dataPath)
//		if err != nil {
//			return err
//		}
//		totalNumber := 0
//		for _, cam := range cameras {
//			if StringIn(cam.Name(), []string{"ZXB_LC03L", "ZXB_LC03M", "ZXB_LC03R"}) {
//				camera := cam.Name()
//				cameraPath := path.Join(dataPath, camera)
//				originPath := cameraPath
//				fs, err := ioutil.ReadDir(originPath)
//				if err != nil {
//					logger.Error(err)
//					continue
//				}
//				for _, f := range fs {
//					if f.Name() == "OrgImg" {
//						originPath = path.Join(originPath, f.Name())
//						break
//					}
//				}
//				var originPoints []PointModel
//				err = ReadSourceDataFromLocalDisk(originPath, &project, &originPoints, sourceData.ScanType)
//				if err != nil {
//					return err
//				}
//				savePath := path.Join(cameraPath, "aligned")
//				camerasContent[camera] = map[string]interface{}{
//					"total":      len(originPoints),
//					"aligned":    0,
//					"status":     "doing",
//					"compressed": 0,
//					"save_path":  savePath,
//				}
//				alignRequest["camera"] = camera
//				alignRequest["cur_root"] = originPath
//				alignRequest["save_path"] = savePath
//				var response map[string]interface{}
//				code, err := HTTPClient("POST", requestURL, alignRequest, &response)
//				if code != 200 || err != nil {
//					logger.Error(response)
//					return err
//				}
//				totalNumber += len(originPoints)
//			}
//		}
//		progressContent["total_number"] = totalNumber
//		progressContent["cameras"] = camerasContent
//		progressContentJson, err := json.Marshal(progressContent)
//		if err != nil {
//			return err
//		}
//		redisContent[contentKey] = progressContentJson
//		// RedisClient.HMSet(redisKey, progressMap)
//		if err = RedisClient.HMSet(progressKey, redisContent).Err(); err != nil {
//			return err
//		}
//	}
//	return nil
//}
