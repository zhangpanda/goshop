package model

import "time"

// Power 权限节点表
type Power struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:权限ID"`
	ParentID  uint      `json:"parent_id" gorm:"index;default:0;comment:上级权限ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:权限名称"`
	Control   string    `json:"control" gorm:"size:128;comment:权限标识(如goods.list)"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	Children  []Power   `json:"children,omitempty" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// RolePower 角色权限关联表
type RolePower struct {
	ID      uint `json:"id" gorm:"primaryKey;comment:记录ID"`
	RoleID  uint `json:"role_id" gorm:"uniqueIndex:idx_rp;not null;comment:角色ID"`
	PowerID uint `json:"power_id" gorm:"uniqueIndex:idx_rp;not null;comment:权限ID"`
}
