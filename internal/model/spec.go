package model

import "time"

// SpecTemplate 规格模板表
type SpecTemplate struct {
	ID        uint       `json:"id" gorm:"primaryKey;comment:模板ID"`
	Name      string     `json:"name" gorm:"size:128;not null;comment:模板名称"`
	Types     []SpecType `json:"types,omitempty" gorm:"foreignKey:TemplateID"`
	CreatedAt time.Time  `json:"created_at"`
}

// SpecType 规格类型表(如颜色/内存)
type SpecType struct {
	ID         uint        `json:"id" gorm:"primaryKey;comment:规格类型ID"`
	TemplateID uint        `json:"template_id" gorm:"index;not null;comment:模板ID"`
	Name       string      `json:"name" gorm:"size:128;not null;comment:规格类型名称(如颜色)"`
	Sort       int         `json:"sort" gorm:"default:0;comment:排序"`
	Values     []SpecValue `json:"values,omitempty" gorm:"foreignKey:TypeID"`
}

// SpecValue 规格值表(如红色/蓝色)
type SpecValue struct {
	ID     uint   `json:"id" gorm:"primaryKey;comment:规格值ID"`
	TypeID uint   `json:"type_id" gorm:"index;not null;comment:规格类型ID"`
	Value  string `json:"value" gorm:"size:128;not null;comment:规格值(如红色)"`
	Sort   int    `json:"sort" gorm:"default:0;comment:排序"`
}
