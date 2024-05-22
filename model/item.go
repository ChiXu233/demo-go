package model

import (
	"github.com/lib/pq"
)

type ItemModel struct {
	Roi          pq.Float64Array `gorm:"column:roi;type:float8[]" json:"-"`
	RoiType      string          `gorm:"column:roi_type" json:"roi_type"`
	RoiCode      string          `gorm:"column:roi_code" json:"roi_code"`
	RoiNumber    int             `gorm:"column:roi_number" json:"roi_number"`
	RoiSource    int             `gorm:"column:roi_source;default:1" json:"roi_source"` //标记来源，1表示灰度图，2表示彩色图
	StandardName string          `gorm:"column:standard_name" json:"standard_name"`
	FourLevelEncode
}

type StandardAndFourLevelEncode struct {
	StandardInfo
	FourLevelEncode
}

type FourLevelEncode struct {
	Name       string `gorm:"column:name" json:"name"`
	Area       string `gorm:"column:area" json:"area"`
	Component  string `gorm:"column:component" json:"component"`
	DetType    string `gorm:"column:det_type" json:"det_type"`
	ErrorTypes string `gorm:"column:error_types" json:"error_types"`
}
type LabelMeJson struct {
	Version string  `json:"version"`
	Shapes  []Shape `json:"shapes"`
}
type Shape struct {
	Label     string      `json:"label"`
	Points    [][]float64 `json:"points"`
	GroupID   int         `json:"group_id"`
	ShapeType string      `json:"shape_type"`
}

type FrontItem struct {
	ID              uint        `json:"ID"`
	Roi             [][]float64 `json:"roi"`
	RoiType         string      `json:"roi_type"`
	Comment         string      `json:"comment"`
	StandardGroupID string      `json:"standard_group_id"`
	ItemModel
}

type MiniItem struct {
	ID        uint        `json:"ID"`
	RoiNumber int         `json:"roi_number"`
	ROI       [][]float64 `json:"roi"`
	ROIType   string      `json:"roi_type"`
	ROISource int         `json:"roi_source"`
	ScanType  string      `json:"scan_type"`
}
