package model

import "gorm.io/gorm"

type Info struct {
	//关联信息
	ID uint `gorm:"primarykey" json:"id"`
	BaseModel
	RelatedInfoModel
	//基本信息
	InfoModel
	InfoModelExtend
	InfoModelStandard
	GroupNumber     uint   `gorm:"column:group_number" json:"group_number"`         // 分组名称
	AlgorithmResult string `gorm:"column:algorithm_result" json:"algorithm_result"` // 算法检测结果
	IsDuplicated    string `gorm:"column:is_duplicated" json:"is_duplicated"`       // 是否被覆盖
	IsMatchSource   string `gorm:"column:is_match_source" json:"is_match_source"`   // 是否是匹配源（标准图）
	IsRecheck       string `gorm:"column:is_recheck" json:"is_recheck"`             // 待明确
	ReferenceStatus string `gorm:"column:reference_status;default false" json:"reference_status"`
	Processed       bool   `gorm:"column:processed;default false" json:"processed"`
}

func (Info) TableName() string {
	return "info"
}

func (info Info) GetLatestOptLog(tx *gorm.DB, operations ...string) (InfoPolyLog, error) {
	var log InfoPolyLog
	selector := make(map[string]interface{})
	selector["image_id"] = info.ImageID
	selector["operation IN"] = operations
	selector["source_data_path"] = info.SourceDataPath
	selector["order"] = "id desc"
	err := QueryEntityByFilter(&selector, &log, tx)
	if err != nil {
		return log, err
	}
	return log, nil
}

type TemporaryInfo struct {
	BaseModel
	ID uint `gorm:"primarykey" json:"id"`
	//关联信息
	RelatedInfoModel
	//基本信息
	InfoModel
	InfoModelExtend
	StandardName string `gorm:"column:standard_name" json:"standard_name"`
}

func (TemporaryInfo) TableName() string {
	return "temporary_info"
}

type RelatedInfoModel struct {
	ProjectID uint   `gorm:"column:project_id" json:"project_id"`
	ScanType  string `gorm:"column: scan_type" json:"scan_type"`
	ParseID   uint   `gorm:"column:parse_id" json:"parse_id"`
	RuleID    uint   `gorm:"column:rule_id" json:"rule_id"`
}

type InfoSimple struct {
	ID uint `json:"id"`
	InfoModel
	CandidateStandard string `json:"candidate_standard"`
	PointName         string `json:"point_name"`
	IsFault           bool   `json:"is_fault"`
	IsStandard        bool   `json:"is_standard"`
	ISChecked         bool   `json:"is_checked"`
}

type SourceDataPathInfo struct {
	ID               uint   `json:"id"`
	ImageID          string `json:"name"`
	ImageURL         string `json:"image_url"`
	ImageURLCompress string `json:"image_url_compress"`
	InfoModelExtend
}

type StandardInfoSimple struct {
	ID uint `json:"id"`
	InfoModel
}

type InfoModel struct {
	ImageID          string `gorm:"column:image_id" json:"image_id"`
	ImageURL         string `gorm:"column:image_url" json:"image_url"`
	ImageURLCompress string `gorm:"column:image_url_compress" json:"image_url_compress"`
}

type InfoModelExtend struct {
	DepthURL        string `gorm:"column:depth_url" json:"depth_url"`
	DepthRenderURL  string `gorm:"column:depth_render_url" json:"depth_render_url"`
	PointCloud      string `gorm:"column:point_cloud" json:"point_cloud"`
	RGBURL          string `gorm:"column:rgb_url" json:"rgb_url"`
	Texture16bitURL string `gorm:"column:texture_16bit_url" json:"texture_16bit_url"`
	DebugTextureURL string `gorm:"column:debug_texture_url" json:"debug_texture_url"`
	SourceDataPath  string `gorm:"column:source_data_path" json:"source_data_path"`
	Comment         string `gorm:"column:comment" json:"comment"`
}

type InfoModelStandard struct {
	StandardID              uint   `gorm:"column:standard_id" json:"standard_id"`
	StandardImageID         string `gorm:"column:standard_image_id" json:"standard_image_id"`
	StandardName            string `gorm:"column:standard_name" json:"standard_name"`
	StandardTexture         string `gorm:"column:standard_texture" json:"standard_texture"`
	StandardTextureCompress string `gorm:"column:standard_texture_compress" json:"standard_texture_compress"`
	StandardDepth           string `gorm:"column:standard_depth" json:"standard_depth"`
	StandardGroupID         string `gorm:"column:standard_group_id" json:"standard_group_id"`
}

type GroupInfo struct {
	ID uint `json:"id"`
	InfoModel
	IsStandard   bool `json:"is_standard"`
	IsFault      bool `json:"is_fault"`
	IsDuplicated bool `json:"is_duplicate"`
	IsRecheck    bool `json:"is_recheck"`
}

type InfoAndItem struct {
	ID           uint   `json:"id"`
	StandardName string `json:"standard_name"`
	ImageUrl     string `json:"image_url"`
	FourLevelEncode
}

type ReservedInfo struct {
	ID               uint   `json:"id"`
	ImageID          string `json:"image_id"`
	ImageURL         string `json:"image_url"`
	GroupName        uint   `json:"group_name"`
	StandardID       uint   `json:"standard_id"`
	StandardImageID  string `json:"standard_image_id"`
	StandardImageURL string `json:"standard_image_url"`
}

// ReturnStruct 检查当前是否有用户在上传数据
type ReturnStruct struct {
	UploadStatus bool      `json:"upload_status"`
	Message      string    `json:"message"`
	MatchRule    MatchRule `json:"match_rule"`
	Ratio        uint      `json:"ratio"`
	IsDisplay    bool      `json:"is_display"`
}

// RelateToRequestStruct 匹配结果关联至请求结构体
type RelateToRequestStruct struct {
	InfoIDList []uint `json:"info_id_list"`
	Operation  string `json:"operation"`
	StandardID uint   `json:"standard_id"`
	SearchName string `json:"search_name"`
}

// RelateToResponseStruct 匹配结果关联至返回结构体
type RelateToResponseStruct struct {
	InfoList          []Info          `json:"info_list"`
	StandardInfoList  []StandardInfo  `json:"standard_info_list"`
	TemporaryInfoList []TemporaryInfo `json:"temporary_info_list"`
	Message           string          `json:"message"`
}

type InfoAndItems struct {
	ID                uint              `json:"id"`
	StandardName      string            `json:"standard_name"`
	ImageURL          string            `json:"image_url"`
	ImageChangeStatus bool              `json:"image_change_status"`
	FourLevel         []FourLevelEncode `json:"four_level"`
}

type MatchSourceResults struct {
	ID uint `json:"id"`
	// RuleID           uint   `json:"rule_id"`
	ImageUrl         string `json:"image_url"`
	ImageUrlCompress string `json:"image_url_compress"`
	GroupNumber      uint   `json:"group_number"`
	IsAside          bool   `json:"is_aside"`
}

type StandardInformation struct {
	ImageFileName string `json:"image_file_name"`
	ImageID       string `json:"image_id"`
	StandardName  string `json:"standard_name"`
	FilePath      string `json:"file_path"`
	JsonPath      string `json:"json_path"`
	Camera        string `json:"camera"`
	Deployment    string `json:"deployment"`
	TrainType     string `json:"train_type"`
	ScanID        string `json:"scan_id"`
}

// InfoPolyLog 匹配关系聚合样本数据操作记录
type InfoPolyLog struct {
	BaseModel
	RelatedInfoModel
	ImageID        string `gorm:"column:image_id" json:"image_id"`                 // 图片id
	SourceDataPath string `gorm:"column:source_data_path" json:"source_data_path"` // 匹配规则样本数据组路径
	SourceDataName string `gorm:"column:source_data_name" json:"source_data_name"` // 聚合样本数据组名称
	BakDataPath    string `gorm:"column:bak_data_path" json:"-"`                   // 备份样本数据  用于delete删除后回滚操作
	Operation      string `gorm:"column:operation" json:"operation"`               // 操作 add、delete、replace
}

func (log InfoPolyLog) TableName() string {
	return "info_poly_log"
}
