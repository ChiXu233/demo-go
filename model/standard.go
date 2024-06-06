package model

import (
	"github.com/wonderivan/logger"
	"strings"
	"time"

	"gorm.io/gorm"
)

type StandardInfo struct {
	ID uint `gorm:"primarykey" json:"id"`
	BaseModel
	ProjectID uint   `gorm:"column:project_id" json:"project_id"`
	ScanType  string `gorm:"column:scan_type;default:'accurate'" json:"scan_type"`
	InfoModel
	InfoModelExtend
	ParseID             uint   `gorm:"column:parse_id" json:"parse_id"`
	RuleID              uint   `gorm:"column:rule_id" json:"rule_id"`
	StandardName        string `gorm:"column:standard_name" json:"standard_name"`
	GroupNumber         uint   `gorm:"column:group_number;force" json:"group_number"`
	CheckingStatus      int    `gorm:"column:checking_status; default 0" json:"checking_status"`
	RunningState        int    `gorm:"column:running_state; default 0;force" json:"running_state"`
	AnnotateStatus      string `gorm:"column:annotate_status" json:"-"`
	ReferenceStatus     string `gorm:"column:reference_status;default false" json:"reference_status"`
	DepthRenderURL      string `gorm:"column:depth_render_url" json:"depth_render_url"`
	PointCloudURL       string `gorm:"column:point_cloud_url" json:"point_cloud_url"`
	TrainType           string `gorm:"column:train_type" json:"train_type"`
	Camera              string `gorm:"column:camera" json:"camera"`
	ScanId              string `gorm:"column:scan_id" json:"scan_id"`
	IsAside             bool   `gorm:"column:is_aside" json:"is_aside"`
	DisplayFor3D        bool   `gorm:"column:display_for_3d" json:"display_for_3d"`
	AnnotatedFor3D      bool   `gorm:"column:annotated_for_3d;default:false;force" json:"annotated_for_3d"`
	CommentFor3D        string `gorm:"column:comment_for_3d" json:"comment_for_3d"`
	ConfigStatusForBolt string `gorm:"column:config_status_for_bolt;default:'none'" json:"config_status_for_bolt"`
	ImageQualityForBolt bool   `gorm:"column:image_quality_for_bolt; default:false" json:"image_quality_for_bolt"`
	DisplayForBolt      bool   `gorm:"column:display_for_bolt; default:true " json:"display_for_bolt"`
	// 这个需要全图渲染图，　单个螺栓渲染图没用先注释掉了
	// RenderedImageForBolt string `gorm:"column:rendered_image_for_bolt" json:"rendered_image_for_bolt"`
	LatestBrightness   int    `gorm:"column:latest_brightness;default:500" json:"latest_brightness"`
	RelatedTrainNumber string `gorm:"column:related_train_number;default:''" json:"related_train_number"` //关联车辆编号，/分割，优先级高于项目的车辆编号
	ImageChangeStatus  bool   `gorm:"column:image_change_status;default:false" json:"image_change_status"`
}

func (*StandardInfo) TableName() string {
	return "standard_info"
}

func (t *StandardInfo) BeforeCreate(tx *gorm.DB) error {
	t.DisplayFor3D = true
	return nil
}

type Item struct {
	BaseModel
	ProjectID uint   `gorm:"column:project_id" json:"project_id"`
	ScanType  string `gorm:"column:scan_type;default:'accurate'" json:"scan_type"`
	// PointID 关联标准图表 (standard_info)
	PointID uint `gorm:"column:point_id" json:"point_id"`
	// InfoID 关联参考图表 (reference_info)
	InfoID          uint   `gorm:"column:info_id" json:"-"`
	Enable          int8   `gorm:"column:enable;default:1;force" json:"-"`
	Comment         string `gorm:"column:comment" json:"comment"`
	StandardGroupID string `gorm:"column:standard_group_id" json:"standard_group_id"`
	ErrorTypesOld   string `gorm:"-" json:"-"`
	Position        string `gorm:"column:position" json:"position"`
	ItemModel
	// BoltMark int8 `gorm:"column:bolt_mark;default:0;force" json:"bolt_mark"`
	//ScanType   		string 			`gorm:"column:scan_type" json:"scan_type"`
}

func (*Item) TableName() string {
	return "standard_item"
}

func (g *Item) BeforeUpdate(trans *gorm.DB) (err error) {
	var itemBeforeUpdate Item
	// 记录更新前的error_type字段
	if err := trans.Model(g).Select("error_types").Where("ID = ?", g.ID).Find(&itemBeforeUpdate).Error; err != nil {
		logger.Error(err)
		return err
	}
	g.ErrorTypesOld = itemBeforeUpdate.ErrorTypes
	return nil
}

//func (g *Item) AfterUpdate(trans *gorm.DB) (err error) {
//	// fmt.Println("更新后: error_type_old = ", g.ErrorTypesOld)
//	// fmt.Println("更新后: error_type = ", g.ErrorTypes)
//	// 更新后的故障类型(errorTypes)如果比更新前(errorTypesOld)增加了"松动", 则创建螺栓工具配置。相反情况删除螺栓工具配置。
//	if !strings.Contains(g.ErrorTypesOld, "松动") && strings.Contains(g.ErrorTypes, "松动") {
//		var boltConfig BoltConfig
//		boltConfig.BoltImageID = g.PointID
//		boltConfig.ItemID = g.ID
//		err = CreateEntity(DB, &boltConfig)
//		if err != nil {
//			return err
//		}
//	}
//	if strings.Contains(g.ErrorTypesOld, "松动") && !strings.Contains(g.ErrorTypes, "松动") {
//		return trans.Where("item_id = ?", g.ID).Delete(&BoltConfig{}).Error
//	}
//	return err
//}

// AfterCreate 标准图的框的新增和删除，同时影响螺栓工具的配置的新增和删除
func (g *Item) AfterCreate(trans *gorm.DB) (err error) {
	// pointID = 0 表示是参考图
	if g.PointID == 0 {
		return
	}
	if strings.Contains(g.DetType, "ignore") {
		return
	}
	// if !strings.Contains(g.DetType, "螺") && !strings.Contains(g.DetType, "油孔") {
	// 	return
	// }
	if !strings.Contains(g.ErrorTypes, "松动") {
		return
	}
	// 暂时先查出来对应的标准图．实际上螺栓工具配置是可以不包含标准图的信息，只关联标注框的，之后先改完螺栓工具的代码，再去掉这里的查询
	// var info StandardInfo
	// err = QueryEntity(g.PointID, &info)
	// if err != nil {
	// 	return err
	// }
	//var boltConfig BoltConfig
	//boltConfig.BoltImageID = g.PointID
	//// boltConfig.BoltImageName = info.ImageID
	//// boltConfig.BoltName = strings.Join([]string{g.Area, g.Component, g.DetType}, "-")
	//// boltConfig.PointBox = g.Roi
	//boltConfig.ItemID = g.ID
	//// boltConfig.DepthImageURL = info.DepthURL
	//// boltConfig.TextureImageURL = info.ImageURL
	//// boltConfig.ImageURLCompress = info.ImageURLCompress
	//err = CreateEntity(DB, &boltConfig)
	return err
}

//func (g *Item) AfterDelete(trans *gorm.DB) (err error) {
//	cond := make(map[string]interface{})
//	cond["item_id"] = g.ID
//	return trans.Delete(&BoltConfig{}, cond).Error
//}

type FrontStandardGroupStructure struct {
	Name             string `json:"name"`
	ProjectID        uint   `json:"project_id"`
	ScanType         string `json:"scan_type"`
	StandardRapidDir string `json:"standard_rapid_dir"`
	Comment          string `json:"comment"`
	PreProcessID     uint   `json:"preprocess_id"`
}

type StandardGroup struct {
	BaseModel
	UpdatedTime    time.Time `gorm:"column:update_time" json:"update_time"`
	Name           string    `gorm:"column:name" json:"name"`
	ProjectInfo    string    `gorm:"column:project" json:"project"`
	Comment        string    `gorm:"column:comment" json:"comment"`
	Numbers        int       `gorm:"column:numbers;force" json:"numbers,omitempty"`
	ScanType       string    `gorm:"column:scan_type" json:"scan_type"`
	ProjectID      uint      `gorm:"column:project_id" json:"project_id"`
	ParseID        uint      `gorm:"column:parse_id" json:"parse_id"`
	UploadStatus   string    `json:"upload_status"`
	UploadProgress int       `json:"upload_progress"`
	PreprocessID   uint      `gorm:"column:preprocess_id" json:"preprocess_id"`
}

func (StandardGroup) TableName() string {
	return "standard_group"
}

// StandardTextureImageCreationHistory 标准图标注灰度图生成调试
type StandardTextureImageCreationHistory struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	SimpleTextureCreationInfo
}

func (StandardTextureImageCreationHistory) TableName() string {
	return "standard_texture_image_creation_history"
}

type SimpleTextureCreationInfo struct {
	StandardID uint `gorm:"column:standard_id" json:"standard_id"`
	BrightNess int  `gorm:"column:brightness" json:"brightness"`
}

type ThreeLevelTrainType struct {
	Company struct {
		Name    string         `json:"name"`
		Project []ProjectLevel `json:"project"`
	} `json:"name"`
}
type ProjectLevel struct {
	Name      string           `json:"name"`
	TrainType []TrainTypeLevel `json:"train_types"`
}
type TrainTypeLevel struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type HdrItem struct {
	Name       string    `json:"name"`
	Area       string    `json:"area"`
	Component  string    `json:"component"`
	DetType    string    `json:"det_type"`
	ErrorTypes string    `json:"error_types"`
	Roi        []float32 `json:"roi"`
}
type ImageInfo struct {
	ID               uint
	ImageUrlCompress string `json:"image_url_compress"`
	PointName        string `json:"point_name"`
	Items            []HdrItem
}
