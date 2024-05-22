package model

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type MatchRule struct {
	gorm.Model
	BaseInfo
	ProjectInfo  string `gorm:"column:project_info" json:"project_info"`
	ParseInfo    string `gorm:"column:parse_info" json:"parse_info"`
	Delimiter    string `gorm:"column:delimiter" json:"delimiter"`
	ImageSuffix  string `gorm:"column:image_suffix" json:"image_suffix"`
	RuleMessage  string `gorm:"column:rule_message" json:"rule_message"`
	FieldLength  uint   `gorm:"column:field_length" json:"field_length"`
	RunningState string `gorm:"column:running_state; default:'false'" json:"running_state"`
	ToBeConfirm  string `gorm:"column:to_be_confirm" json:"to_be_confirm"`
	Operation    string `gorm:"column:operation" json:"operation"`
	KeepHistory  string `gorm:"column:keep_history; default:'false';force" json:"keep_history"`
	RuleType     string `gorm:"column:rule_type;default:base" json:"rule_type"`     //规则类型，默认基本规则，可选：special特殊规则 base基本规则
	RuleMark     string `gorm:"column:rule_mark;default:''" json:"rule_mark"`       //标记特殊结构，从”A“开始
	BaseRuleID   int    `gorm:"column:base_rule_id;default:-1" json:"base_rule_id"` //特殊规则对应的基本规则
}

func (match MatchRule) DataPathIsUpdate(newDataPath []string) bool {
	if len(newDataPath) != len(match.DataPathList) {
		return true
	}
	for idx, dataPath := range newDataPath {
		if match.DataPathList[idx] != dataPath {
			return true
		}
	}
	return false
}

func (MatchRule) TableName() string {
	return "match_rule"
}

type BaseInfo struct {
	Name               string         `gorm:"column:name" json:"name"`
	ProjectID          uint           `gorm:"column:project_id" json:"project_id" form:"project_id"`
	Comment            string         `gorm:"column:comment" json:"comment"`
	DataPathList       pq.StringArray `gorm:"column:data_path_list;type:text[]" json:"data_path_list"`
	PolyDataPathList   pq.StringArray `gorm:"column:poly_data_path_list;type:text[]" json:"poly_data_path_list"` // 聚合样本数据组路径
	ParseID            uint           `gorm:"column:parse_id" json:"parse_id"`
	ImageIDFields      pq.Int64Array  `gorm:"column:image_id_fields;type:bigint[]" json:"image_id_fields"`
	RelatedTrainNumber string         `gorm:"column:related_train_number;default:''" json:"related_train_number"` //关联车辆编号，/分割，优先级高于项目的车辆编号
}

// InitMatchRule 初始化运行状态
func InitMatchRule(DB *gorm.DB) error {
	err := DB.Table("match_rule").Where("deleted_at is null").Update("running_state", "false").Error
	if err != nil {
		return err
	}
	return nil
}

// MatchRuleRequest 匹配结果请求结构体
type MatchRuleRequest struct {
	BaseInfo
	ID          uint   `json:"id"`
	RuleMessage string `json:"rule_message"`
	Operation   string `json:"operation"`
	KeepHistory string `json:"keep_history"`
	Mode        string `json:"mode"`
	DataMode    string `json:"data_mode"`
	SpecialRule bool   `json:"special_rule"`
	Header      string `json:"header"`
}

// MatchResult 匹配结果
type MatchResult struct {
	GroupNumber uint              `json:"group_number"`
	InfoS       []InfoSimple      `json:"infos"`
	GroupInfo   GroupInfoStandard `json:"group_info"`
}
type GroupInfoStandard struct {
	StandardInfoSimple
	GroupInfoStruct
}
type GroupInfoStruct struct {
	StandardImageID string `json:"standard_image_id"`
	MatchResultCount
}
type MatchResultCount struct {
	Total    uint `json:"total"`
	Pass     uint `json:"pass"`
	Failed   uint `json:"failed"`
	Unknown  uint `json:"unknown"`
	Running  uint `json:"running"`
	Progress uint `json:"progress"`
	Prop     uint `json:"prop"`
}

type ExportStruct struct {
	GroupNumber uint             `json:"group_number"`
	GroupInfo   GroupInfo4Export `json:"group_info"`
	InfoS       []InfoSimple     `json:"infos"`
}

type GroupInfo4Export struct {
	Count     MatchResultCount `json:"count"`
	ID        uint             `json:"id"`
	ImageID   string           `json:"image_id"`
	ImageURL  string           `json:"image_url"`
	PointName string           `json:"point_name"`
}

type MatchRelation struct {
	Info         Info
	StandardInfo StandardInfo
}
