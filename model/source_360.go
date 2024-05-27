package model

import (
	"demo-go/config"
	. "demo-go/utils"
	"fmt"
	"github.com/lib/pq"
	"github.com/wonderivan/logger"
	"gorm.io/gorm"
	"math"
	"time"
)

// BaseModel1 和BaseModel相比，减少update at, delete at的json返回
type BaseModel1 struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type PreProcess struct {
	BaseModel1
	ProjectID   uint   `gorm:"column:project_id" json:"project_id"`
	ScanType    string `gorm:"column:scan_type" json:"scan_type"`
	ProjectInfo string `gorm:"column:-" json:"project_info"`
	Name        string `gorm:"column:name" json:"name"`
	Comment     string `gorm:"column:comment" json:"comment"`
}

func (*PreProcess) TableName() string {
	return "pre_process"
}

func (p *PreProcess) AfterFind(*gorm.DB) error {
	var err error
	var project Project
	err = QueryEntity(p.ProjectID, &project)
	if err != nil {
		return err
	}
	p.ProjectInfo = project.Info()
	return nil
}

type Source360 struct {
	BaseModel1
	PreProcessID uint           `gorm:"column:preprocess_id" json:"preprocess_id"`
	Name         string         `gorm:"column:name" json:"name"`
	Cameras      pq.StringArray `gorm:"column:cameras;type:text[];" json:"cameras"`
	Direction    string         `gorm:"column:direction" json:"direction"`
	DataPath     string         `gorm:"column:data_path" json:"data_path"`
	Status       string         `gorm:"column:status" json:"status"`
	Progress     float64        `gorm:"-" json:"progress"`
}

func (*Source360) TableName() string {
	return "source_360"
}

func (source *Source360) AfterFind(*gorm.DB) error {
	var err error
	if source.Status == "none" {
		source.Progress = 0
		return nil
	}
	if source.Status == "finished" {
		source.Progress = 100
		return nil
	}
	if source.Status == "ongoing" {
		sourceCreateRedisKey := "create_source_progress_" + String(config.Conf.APP.IP) + String(config.Conf.APP.Port)
		contentKey := "source_" + String(source.ID)
		if !RedisHExists(sourceCreateRedisKey, contentKey) {
			source.Status = "failed"
			source.Progress = 0
			logger.Info("key %s %s not exists", sourceCreateRedisKey, contentKey)
		}
		var progressContent map[string]interface{}
		_, err = QueryRedisProgress(sourceCreateRedisKey, contentKey, &progressContent)
		if err != nil {
			source.Status = "failed"
			source.Progress = 0
			logger.Info(err)
		}
		total := progressContent["total"]
		processed := progressContent["processed"]
		var progress float64 = 0
		if total.(float64) != 0 {
			progress = math.Ceil((processed.(float64) / total.(float64)) * 100)
		}
		source.Progress = progress
	}
	return nil
}

type Source360Image struct {
	BaseModel1
	SourceID                 uint   `gorm:"column:source_id" json:"-"`
	Name                     string `gorm:"column:name" json:"name"`
	Number                   int    `gorm:"column:number" json:"-"`
	Path                     string `gorm:"column:path" json:"-"`
	PathCompressed           string `gorm:"column:path_compressed" json:"-"`
	PathCompressedLowQuality string `gorm:"column:path_compressed_low_quality" json:"-"`
	URL                      string `gorm:"-" json:"url"`
	URLLowQuality            string `gorm:"-" json:"url_low_quality"`
}

func (*Source360Image) TableName() string {
	return "source_360_image"
}

func (img *Source360Image) AfterFind(*gorm.DB) error {
	returnPath := ""
	if img.PathCompressed != "" {
		returnPath = img.PathCompressed
	} else if img.Path != "" {
		returnPath = img.Path
	}
	if returnPath != "" {
		img.URL = fmt.Sprintf("http://%s:%d/%s", config.Conf.APP.IP, config.Conf.APP.Port, returnPath)
	}
	if img.PathCompressedLowQuality != "" {
		img.URLLowQuality = fmt.Sprintf("http://%s:%d/%s", config.Conf.APP.IP, config.Conf.APP.Port, img.PathCompressedLowQuality)
	}
	return nil
}

type Source360Concat struct {
	BaseModel1
	PreProcessID        uint    `gorm:"column:preprocess_id" json:"-"`
	SourceID            uint    `gorm:"column:source_id" json:"source_id"`
	Name                string  `gorm:"column:name" json:"name"`
	Size                string  `gorm:"column:size" json:"size"`
	Status              string  `gorm:"column:status" json:"status"`
	Progress            float64 `gorm:"-" json:"progress"`
	ImagePath           string  `gorm:"image_path" json:"-"`
	ImagePathLowQuality string  `gorm:"image_path_low_quality" json:"-"`
	ImageURL            string  `gorm:"-" json:"image_url"`
	ImageURLCompressed  string  `gorm:"-" json:"image_url_compressed"`
	Start               float64 `gorm:"column:start" json:"start"`
	End                 float64 `gorm:"column:end" json:"end"`
	Changed             bool    `gorm:"column:changed" json:"changed"`
}

func (*Source360Concat) TableName() string {
	return "source_360_concat"
}

func (concat *Source360Concat) AfterFind(*gorm.DB) error {
	var err error
	if concat.ImagePath != "" {
		concat.ImageURL = fmt.Sprintf("http://%s:%d/%s", config.Conf.APP.IP, config.Conf.APP.Port, concat.ImagePath)
	}
	if concat.ImagePathLowQuality != "" {
		concat.ImageURLCompressed = fmt.Sprintf("http://%s:%d/%s", config.Conf.APP.IP, config.Conf.APP.Port, concat.ImagePathLowQuality)
	}

	if concat.Status == "none" {
		concat.Progress = 0
		return nil
	}
	if concat.Status == "finished" {
		concat.Progress = 100
		return nil
	}
	if concat.Status == "ongoing" {
		sourceCreateRedisKey := "source_concat_progress_" + String(config.Conf.APP.IP) + String(config.Conf.APP.Port)
		contentKey := "source_concat_" + String(concat.ID)
		if !RedisHExists(sourceCreateRedisKey, contentKey) {
			concat.Status = "failed"
			concat.Progress = 0
			logger.Info("key %s %s not exists", sourceCreateRedisKey, contentKey)
			return nil
		}
		var progressContent map[string]interface{}
		_, err = QueryRedisProgress(sourceCreateRedisKey, contentKey, &progressContent)
		if err != nil {
			concat.Status = "failed"
			concat.Progress = 0
			logger.Info(err)
			return nil
		}
		total := progressContent["total"]
		processed := progressContent["processed"]
		var progress float64 = 0
		if total.(float64) != 0 {
			progress = math.Ceil((processed.(float64) / total.(float64)) * 100)
		}
		concat.Progress = progress
	}
	return nil
}

type Source360Item struct {
	BaseModel1
	ConcatID    uint            `gorm:"column:concat_id" json:"concat_id"`
	Type        string          `gorm:"column:type" json:"type"`
	Name        string          `gorm:"column:name" json:"name"`
	ErrorTypes  string          `gorm:"column:error_types" json:"error_types"`
	RoiType     string          `gorm:"column:roi_type" json:"roi_type"`
	Roi         pq.Float64Array `gorm:"column:roi;type:float8[]" json:"-"`
	RoiCode     string          `gorm:"column:roi_code" json:"roi_code"`
	RoiFrontend [][]float64     `gorm:"-" json:"roi"`
	UUID        string          `gorm:"column:uuid" json:"uuid"`
}

func (*Source360Item) TableName() string {
	return "source_360_item"
}

func (item *Source360Item) AfterFind(*gorm.DB) error {
	if len(item.Roi) > 1 {
		for i := 1; i < len(item.Roi); i = i + 2 {
			item.RoiFrontend = append(item.RoiFrontend, []float64{item.Roi[i-1], item.Roi[i]})
		}
	}
	return nil
}

func (item *Source360Item) TransRoi() {
	var roi []float64
	for _, roiFront := range item.RoiFrontend {
		roi = append(roi, roiFront...)
	}
	item.Roi = roi
	return
}

type Source360Rule struct {
	BaseModel1
	GroupID    uint   `gorm:"column:group_id" json:"-"`
	Camera     string `gorm:"column:camera" json:"camera"`
	ConcatID   uint   `gorm:"column:concat_id" json:"concat_id"`
	ConcatName string `gorm:"column:concat_name" json:"concat_name"`
	Type       string `gorm:"column:type" json:"type"` // 生成方式
	Height     int    `gorm:"column:height" json:"height"`
	Overlap    int    `gorm:"column:overlap" json:"overlap"`
	Enable     bool   `gorm:"column:enable" json:"enable"`
}

func (*Source360Rule) TableName() string {
	return "source_360_rule"
}

func (r *Source360Rule) AfterFind(*gorm.DB) error {
	var err error
	var concat Source360Concat
	if r.ConcatID != 0 {
		err = QueryEntity(r.ConcatID, &concat)
		if err != nil {
			return err
		}
		r.ConcatName = concat.Name
	}
	return nil
}
