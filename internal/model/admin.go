package model

import "time"

type Admin struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:管理员ID"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:64;not null;comment:管理员用户名"`
	Password  string    `json:"-" gorm:"size:128;not null;comment:密码(bcrypt)"`
	Nickname  string    `json:"nickname" gorm:"size:64;comment:昵称"`
	RoleID    uint      `json:"role_id" gorm:"index;comment:角色ID"`
	Role      *Role     `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1正常"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:角色ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:角色名称"`
	Desc      string    `json:"desc" gorm:"size:255;comment:角色描述"`
	Powers    string    `json:"powers" gorm:"type:text;comment:权限标识JSON数组"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdminOperationLog 管理员操作审计日志表
type AdminOperationLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:日志ID"`
	AdminID   uint      `json:"admin_id" gorm:"index;not null;comment:管理员ID"`
	Username  string    `json:"username" gorm:"size:64;comment:管理员用户名"`
	Action    string    `json:"action" gorm:"size:128;comment:操作动作"`
	Method    string    `json:"method" gorm:"size:10;comment:请求方法"`
	Path      string    `json:"path" gorm:"size:255;comment:请求路径"`
	IP        string    `json:"ip" gorm:"size:64;comment:请求IP"`
	UserAgent string    `json:"user_agent" gorm:"size:255;comment:浏览器UA"`
	CreatedAt time.Time `json:"created_at"`
}
