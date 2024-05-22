package model

import (
	"fmt"
	"strings"
)

type Project struct {
	BaseModel
	CompanyName    string `gorm:"column:company_name" json:"company_name"`
	Name           string `gorm:"column:name" json:"name"`
	Location       string `gorm:"column:location" json:"location"`
	ScanType       string `gorm:"column:scan_type" json:"scan_type"`
	Comment        string `gorm:"column:comment" json:"comment"`
	HDRProjectID   uint   `gorm:"column:hdr_project_id" json:"hdr_project_id"`
	HDRProjectInfo string `gorm:"column:hdr_project_info" json:"hdr_project_info"`
}

type ProjectIdentifier struct {
	ProjectID uint   `gorm:"column:project_id" json:"project_id"`
	ScanType  string `gorm:"column:scan_type" json:"scan_type"`
}

func (project Project) Info() string {
	return fmt.Sprintf("%s/%s", project.CompanyName, project.Name)
}

func (project Project) ContainScanType(inputScanTypes string) bool {
	for _, inputScanType := range strings.Split(inputScanTypes, "、") {
		if !strings.Contains(project.ScanType, inputScanType) {
			return false
		}
	}
	return true
}

type NewProjectRequest struct {
	ProjectID      uint     `json:"project_id"`
	CompanyName    string   `json:"company_name"`   //合作企业
	Name           string   `json:"name"`           //合作项目
	Location       string   `json:"location"`       //项目地点
	TrainType      string   `json:"train_type"`     //合作车型
	ScanType       string   `json:"scan_type"`      //合作类目
	Comment        string   `json:"comment"`        //备注信息
	TrainSerial    string   `json:"train_serial"`   //列车编号
	HDRProjectID   uint     `json:"hdr_project_id"` //HDR项目ID
	HDRProjectInfo string   `gorm:"column:hdr_project_info" json:"hdr_project_info"`
	DataSource     []string `json:"data_source"`
}

func (Project) TableName() string {
	return "project"
}

type RequestIDList struct {
	IDList []int `json:"id_list"`
}

type ProjectFront struct {
	Project
	ScanID string `json:"scan_id"`
}

var ScanTypeMap = map[string]string{
	"accurate": "精扫",
	"rapid":    "快扫",
	"360":      "360",
}

var ScanTypeMapReverse = map[string]string{
	"精扫":  "accurate",
	"快扫":  "rapid",
	"360": "360",
}

//HDR models

type Project4HDR struct {
	ID          uint
	CompanyID   uint   `json:"company_id"`
	CompanyName string `json:"company_name"`
	Name        string `json:"name"`
}
