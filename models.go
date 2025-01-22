package main

// FgMcatParamsList - структура таблицы fg_mcat_params_list
type FgMcatParamsList struct {
	ID     uint64                `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	Name   string                `gorm:"type:varchar(255);not null;collate:utf8mb4_0900_ai_ci" json:"name"`
	Values []*FgMcatParamsValues `gorm:"many2many:fg_mcat_params_list_values" json:"values"`
}

func (FgMcatParamsList) TableName() string {
	return "fg_mcat_params_list"
}

// FgMcatParamsValues - структура таблицы fg_mcat_params_values
type FgMcatParamsValues struct {
	ID      uint64              `gorm:"primaryKey;autoIncrement;not null" json:"id"`
	ParamID uint64              `gorm:"type:bigint;not null;default:0" json:"param_id"`
	Value   string              `gorm:"type:varchar(255);not null;default:'0';collate:utf8mb4_0900_ai_ci" json:"value"`
	Params  []*FgMcatParamsList `gorm:"many2many:fg_mcat_params_list_values" json:"params"`
}

func (FgMcatParamsValues) TableName() string {
	return "fg_mcat_params_values"
}

// FgMcatParamsListValues - промежуточная таблица для связи многие ко многим
type FgMcatParamsListValues struct {
	FgMcatParamsListID   uint64 `gorm:"not null"`
	FgMcatParamsValuesID uint64 `gorm:"not null"`
}

// FgMcatItems - структура таблицы fg_mcat_items
type FgMcatItems struct {
	IdHash                     string  `gorm:"primaryKey;type:varchar(255);not null;default:'';collate:utf8mb3_bin" json:"id_hash"`
	ItemName                   string  `gorm:"type:text;default:null;collate:utf8mb3_bin" json:"item_name"`
	BrandId                    *uint   `gorm:"type:int unsigned;default:null" json:"brand_id"`
	ItemArticle                string  `gorm:"type:varchar(150);not null;default:'0';collate:utf8mb3_bin" json:"item_article"`
	ItemDescription            string  `gorm:"type:text;default:null;collate:utf8mb3_bin" json:"item_description"`
	ItemDescriptionDisplayType string  `gorm:"type:ENUM('CONCAT', 'OVERRIDE');default:null;collate:utf8mb3_bin" json:"item_description_display_type"`
	ItemMetaTitle              string  `gorm:"type:text;default:null;collate:utf8mb3_bin" json:"item_meta_title"`
	ItemMetaKeywords           string  `gorm:"type:text;default:null;collate:utf8mb3_bin" json:"item_meta_keywords"`
	ItemMetaDescription        string  `gorm:"type:text;default:null;collate:utf8mb3_bin" json:"item_meta_description"`
	ItemLastEditDate           *string `gorm:"type:date;default:null" json:"item_last_edit_date"` // Используйте *string для даты, если она может быть NULL
	IsOriginal                 *uint8  `gorm:"type:tinyint unsigned;default:null" json:"is_original"`
}

func (FgMcatItems) TableName() string {
	return "fg_mcat_items"
}

type FgMcatParams struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement;notNull" json:"id"`
	ItemHash   string `gorm:"type:varchar(255);not null;default:'0';collate:utf8mb3_bin" json:"item_hash"`
	ParamName  string `gorm:"column:ParamName;type:varchar(50);not null;default:'';collate:utf8mb3_bin" json:"param_name"`
	ParamValue string `gorm:"column:ParamValue;type:varchar(50);not null;default:'';collate:utf8mb3_bin" json:"param_value"`
}

func (FgMcatParams) TableName() string {
	return "fg_mcat_params"
}
