package model

type Point struct {
	BaseModel
	SourceDataID       uint   `gorm:"column:source_data_id"`
	StandardName       string `gorm:"column:standard_name" json:"-"`                           // 标准图名称
	ImageURL           string `gorm:"column:image_url" json:"-"`                               // 灰度图URL
	RGBURL             string `gorm:"column:rgb_url" json:"rgb_url"`                           // RGB图URL
	PointCloudURL      string `gorm:"column:point_cloud_url" json:"point_cloud_url"`           // 点云图URL
	CompressedImageURL string `gorm:"column:compressed_image_url" json:"compressed_image_url"` // 压缩灰度图URL
	DepthURL           string `gorm:"column:depth_url" json:"depth_url"`                       // 深度图URL
	RenderedDepthURL   string `gorm:"column:rendered_depth_url" json:"rendered_depth_url"`     // 深度渲染图URL
	Status             string `gorm:"column:status default:normal" json:"status"`              // 是否故障数据
	ErrorTypes         string `gorm:"column:error_types" json:"-"`                             // 包含的故障类型
	PointModel
}

type PointModel struct {
	PointName       string `gorm:"column:point_name" json:"point_name"` // 点位名称
	ImagePath       string `gorm:"column:image_path" json:"-"`          // 灰度图路径
	DepthPath       string `gorm:"column:depth_path" json:"-"`          // 深度图路径
	PointCloudPath  string `gorm:"column:point_cloud_path" json:"-"`    // 点云路径
	RGBPath         string `gorm:"column:rgb_path" json:"-"`            // RGB图路径
	ScanID          string `gorm:"column:scan_id" json:"-"`             // 快扫的camera + scan_id
	IDConnectSymbol string `gorm:"-" json:"-"`
	DataSource      string `gorm:"default:'local'" json:"data_source"`
}
