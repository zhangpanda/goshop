package model

import "time"

// GoodsParamsTemplate 参数模板表
type GoodsParamsTemplate struct {
	ID        uint                `json:"id" gorm:"primaryKey;comment:模板ID"`
	Name      string              `json:"name" gorm:"size:64;not null;comment:模板名称"`
	Configs   []GoodsParamsConfig `json:"configs,omitempty" gorm:"foreignKey:TemplateID"`
	CreatedAt time.Time           `json:"created_at"`
}

// GoodsParamsConfig 参数模板配置项表
type GoodsParamsConfig struct {
	ID         uint   `json:"id" gorm:"primaryKey;comment:配置ID"`
	TemplateID uint   `json:"template_id" gorm:"index;not null;comment:模板ID"`
	Name       string `json:"name" gorm:"size:64;not null;comment:参数名"`
	Value      string `json:"value" gorm:"size:255;comment:参数默认值"`
	Sort       int    `json:"sort" gorm:"default:0;comment:排序"`
}

// GoodsParams 商品参数值表
type GoodsParams struct {
	ID      uint   `json:"id" gorm:"primaryKey;comment:参数ID"`
	GoodsID uint   `json:"goods_id" gorm:"index;not null;comment:商品ID"`
	Name    string `json:"name" gorm:"size:64;not null;comment:参数名"`
	Value   string `json:"value" gorm:"size:255;comment:参数值"`
	Sort    int    `json:"sort" gorm:"default:0;comment:排序"`
}
