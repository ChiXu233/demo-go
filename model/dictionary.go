package model

type Dict struct {
	BaseModel
	DictName    string `gorm:"column:dict_name" json:"dict_name"`
	ProjectID   uint   `gorm:"column:project_id" json:"project_id"`
	ProjectInfo string `gorm:"column:project_info" json:"project_info"`
	ParseID     uint   `gorm:"column:parse_id" json:"parse_id"`
	ParseInfo   string `gorm:"column:parse_info" json:"parse_info"`
	Creator     string `gorm:"column:creator" json:"creator"`
	ScanType    string `gorm:"column:scan_type" json:"scan_type"`
	Comment     string `gorm:"column:comment" json:"comment"`
}

func (Dict) TableName() string {
	return "dict"
}

type DictItem struct {
	BaseModel
	BaseItem
	CodeSource int `gorm:"column:code_source;default 1" json:"code_source"` //1表示来自图片ID，2表示来自文件夹
}

func (DictItem) TableName() string {
	return "dict_item"
}

type BaseItem struct {
	DictID       uint   `gorm:"dict_id" json:"dict_id"`
	CategoryName string `gorm:"category_name" json:"category_name"`
	Code         string `gorm:"code" json:"code"`
	CodeName     string `gorm:"code_name" json:"code_name"`
	Comment      string `gorm:"column:comment" json:"comment"`
}
type NewDictRequest struct {
	ID        uint   `json:"id"`
	DictName  string `json:"dict_name"`
	ProjectId uint   `json:"project_id"`
	ParseID   uint   `json:"parse_id"`
	// RelatedInfo string `json:"related_info"`
	Comment string `json:"comment"`
}

type DictList struct {
	Order uint `json:"order"`
	Dict
}

type DictItemList struct {
	BaseModel
	BaseItem
	Order      uint   `json:"order"`
	CodeSource string `json:"code_source"` //1表示来自图片ID，2表示来自文件夹

}
type DictItemListRequest struct {
	IDList           []uint   `json:"id_list"`
	CategoryNameList []string `json:"category_name_list"`
	CodeList         []string `json:"code_list"`
	CodeNameList     []string `json:"code_name_list"`
	CommentList      []string `json:"comment_list"`
}

type NewDictItemRequest struct {
	CategoryName string `json:"category_name"`
	Code         string `json:"code"`
	CodeName     string `json:"code_name"`
	Comment      string `json:"comment"`
}

var CodeSourceMap = map[int]string{
	1: "图片",
	2: "数据组",
}

type DictItemInfo struct {
	BaseModel
	BaseItem
	CodeSource string `json:"code_source"` //1表示来自图片ID，2表示来自文件夹
}
