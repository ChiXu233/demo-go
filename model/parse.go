package model

type Parse struct {
	BaseModel
	ParseName                string `gorm:"column:parse_name" json:"parse_name"`
	ProjectID                uint   `gorm:"column:project_id" json:"project_id"`
	ImageID                  string `gorm:"column:image_id" json:"image_id"`
	ImageIDParse             string `gorm:"column:image_id_parse" json:"image_id_parse"`
	IDConnectSymbol          string `gorm:"column:id_connect_symbol" json:"id_connect_symbol"`
	OtherConnectSymbol       string `gorm:"column:other_connect_symbol" json:"other_connect_symbol"`
	ContinueSymbol           bool   `gorm:"column:continue_symbol; default:true" json:"continue_symbol" `
	ImageSuffix              string `gorm:"column:image_suffix" json:"image_suffix"`
	IDIndex                  uint   `gorm:"column:id_index" json:"id_index"`
	Comment                  string `gorm:"column:comment" json:"comment"`
	ProjectInfo              string `gorm:"column:project_info" json:"project_info"`
	FileDir                  string `gorm:"column:file_dir" json:"file_dir"`
	FileDirParse             string `gorm:"column:file_dir_parse" json:"file_dir_parse"`
	IDConnectSymbolForDir    string `gorm:"column:id_connect_symbol_for_dir" json:"id_connect_symbol_for_dir"`
	OtherConnectSymbolForDir string `gorm:"column:other_connect_symbol_for_dir" json:"other_connect_symbol_for_dir"`
	ContinueSymbolForDir     bool   `gorm:"column:continue_symbol_for_dir; default:true" json:"continue_symbol_for_dir" `
	IDIndexForDir            uint   `gorm:"column:id_index_for_dir" json:"id_index_for_dir"`
}

func (Parse) TableName() string {
	return "parse"
}

type SimpleParse struct {
	ID        uint   `json:"ID"`
	ParseName string `json:"parse_name"`
}
type ParseDecode struct {
	BaseModel
	ProjectID   uint   `gorm:"column:project_id" json:"project_id"`
	ParseID     uint   `gorm:"column:parse_id" json:"parse_id"`
	ImageField  string `gorm:"column:image_field" json:"image_field"`
	ParseCode   uint   `gorm:"column:parse_code" json:"parse_code"`
	ParseName   string `gorm:"column:parse_name" json:"parse_name"`
	DirField    string `gorm:"column:dir_field" json:"dir_field"`
	FieldSource int    `gorm:"column:field_source;default:1" json:"field_source"`
	Delimiter   string `gorm:"column:delimiter" json:"delimiter"`
	NeedEnter   bool   `gorm:"column:need_enter" json:"need_enter"`
	ScanType    string `gorm:"column:scan_type" json:"scan_type"`
}

func (ParseDecode) TableName() string {
	return "parse_decode"
}

type NewParseRequest struct {
	ID        uint   `json:"ID"`
	ParseName string `json:"parse_name"`
	ProjectID uint   `json:"project_id"`
	Comment   string `json:"comment"`

	ImageID            string `json:"image_id"`
	ImageIDParse       string ` json:"image_id_parse"`
	IDConnectSymbol    string `json:"id_connect_symbol"`
	OtherConnectSymbol string `json:"other_connect_symbol"`
	ContinueSymbol     bool   `json:"continue_symbol"`
	IDIndex            uint   `json:"id_index"`

	FileDir                  string `json:"file_dir"`
	FileDirParse             string ` json:"file_dir_parse"`
	IDConnectSymbolForDir    string `json:"id_connect_symbol_for_dir"`
	OtherConnectSymbolForDir string `json:"other_connect_symbol_for_dir"`
	ContinueSymbolForDir     bool   `json:"continue_symbol_for_dir"`
	IDIndexForDir            uint   `json:"id_index_for_dir"`
	Option                   string `json:"option"`
}

type ParseWithImageDecode struct {
	Parse
	ParseDecode ParseImageID `json:"parse_decode"`
}

type ParseDetails struct {
	Parse
	ImageIDParseResult ParseResult `json:"image_id_parse_result"`
	FileDirParseResult ParseResult `json:"file_dir_parse_result"`
}

type ParseBasicInfo struct {
	BaseModel
	ParseName                string `json:"parse_name"`
	ProjectID                uint   ` json:"project_id"`
	Comment                  string `json:"comment"`
	ProjectInfo              string `json:"project_info"`
	IDConnectSymbol          string `gorm:"column:id_connect_symbol" json:"id_connect_symbol"`
	OtherConnectSymbol       string `gorm:"column:other_connect_symbol" json:"other_connect_symbol"`
	ContinueSymbol           string `gorm:"column:continue_symbol; default:'true''" json:"continue_symbol" `
	ImageSuffix              string `gorm:"column:image_suffix" json:"image_suffix"`
	IDIndex                  uint   `gorm:"column:id_index" json:"id_index"`
	IDConnectSymbolForDir    string `gorm:"column:id_connect_symbol_for_dir" json:"id_connect_symbol_for_dir"`
	OtherConnectSymbolForDir string `gorm:"column:other_connect_symbol_for_dir" json:"other_connect_symbol_for_dir"`
	ContinueSymbolForDir     string `gorm:"column:continue_symbol_for_dir; default:'true''" json:"continue_symbol_for_dir" `
	IDIndexForDir            uint   `gorm:"column:id_index_for_dir" json:"id_index_for_dir"`
}
type ParseImageID struct {
	ImageID            string      `json:"image_id"`
	ImageIDParse       string      `json:"image_id_parse"`
	ImageIDParseResult ParseResult `json:"image_id_parse_result"`
}

type ParseFileDir struct {
	FileDir            string      `json:"file_dir"`
	FileDirParse       string      `json:"file_dir_parse"`
	FileDirParseResult ParseResult `json:"file_dir_parse_result"`
}

type ParseResult struct {
	SplitResult []string ` json:"split_result"`
	ParseCode   []uint   `json:"parse_code"`
	ParseName   []string ` json:"parse_name"`
	NeedEnter   []bool   `json:"need_enter"`
}
