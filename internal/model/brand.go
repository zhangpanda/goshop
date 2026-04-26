package model

import "time"

// Brand 品牌表
type Brand struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:品牌ID"`
	Name      string    `json:"name" gorm:"size:100;not null;comment:品牌名称"`
	Logo      string    `json:"logo" gorm:"size:255;comment:品牌Logo"`
	Desc      string    `json:"desc" gorm:"size:500;comment:品牌描述"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
