package model

import "time"

// Category 商品分类表
type Category struct {
	ID        uint       `json:"id" gorm:"primaryKey;comment:分类ID"`
	ParentID  uint       `json:"parent_id" gorm:"index;default:0;constraint:false;comment:上级分类ID(0为顶级)"`
	Name      string     `json:"name" gorm:"size:64;not null;comment:分类名称"`
	Icon      string     `json:"icon" gorm:"size:255;comment:分类图标"`
	Sort      int        `json:"sort" gorm:"default:0;comment:排序(越大越前)"`
	Status    int8       `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	Children  []Category `json:"children,omitempty" gorm:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
