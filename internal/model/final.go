package model

import "time"

// Design 页面设计模板表
type Design struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:设计ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:设计名称"`
	Data      string    `json:"data" gorm:"type:longtext;comment:设计配置JSON"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FormTableUserFields 表单自定义字段表
type FormTableUserFields struct {
	ID     uint   `json:"id" gorm:"primaryKey;comment:字段ID"`
	FormID uint   `json:"form_id" gorm:"index;not null;comment:表单ID"`
	Name   string `json:"name" gorm:"size:64;not null;comment:字段名称"`
	Type   string `json:"type" gorm:"size:32;comment:字段类型:text/select/radio/checkbox/textarea"`
	Config string `json:"config" gorm:"type:text;comment:字段配置JSON"`
	Sort   int    `json:"sort" gorm:"default:0;comment:排序"`
}

// GoodsContentApp APP端商品详情表
type GoodsContentApp struct {
	ID      uint   `json:"id" gorm:"primaryKey;comment:记录ID"`
	GoodsID uint   `json:"goods_id" gorm:"uniqueIndex;not null;comment:商品ID"`
	Content string `json:"content" gorm:"type:longtext;comment:APP端商品详情"`
}

// Layout 首页布局配置表
type Layout struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:布局ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:布局名称"`
	Type      string    `json:"type" gorm:"size:32;comment:布局类型:home/category/user"`
	Data      string    `json:"data" gorm:"type:longtext;comment:布局配置JSON"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderService 订单客服/服务记录表
type OrderService struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	OrderID   uint      `json:"order_id" gorm:"index;not null;comment:订单ID"`
	UserID    uint      `json:"user_id" gorm:"index;comment:用户ID"`
	AdminID   uint      `json:"admin_id" gorm:"index;comment:管理员ID"`
	Type      string    `json:"type" gorm:"size:32;comment:类型:consult/complaint/other"`
	Content   string    `json:"content" gorm:"type:text;comment:咨询内容"`
	Reply     string    `json:"reply" gorm:"type:text;comment:回复内容"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0待处理1已处理"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PayLogValue 支付日志关联订单明细表
type PayLogValue struct {
	ID       uint  `json:"id" gorm:"primaryKey;comment:记录ID"`
	PayLogID uint  `json:"pay_log_id" gorm:"index;not null;comment:支付日志ID"`
	OrderID  uint  `json:"order_id" gorm:"index;not null;comment:订单ID"`
	Amount   int64 `json:"amount" gorm:"not null;comment:该订单支付金额(分)"`
}

// PluginsDataConfig 插件数据配置表
type PluginsDataConfig struct {
	ID       uint   `json:"id" gorm:"primaryKey;comment:记录ID"`
	PluginID uint   `json:"plugin_id" gorm:"index;not null;comment:插件ID"`
	Key      string `json:"key" gorm:"size:64;not null;comment:配置键"`
	Value    string `json:"value" gorm:"type:text;comment:配置值"`
}

// QuickNav 快捷导航表
type QuickNav struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:导航ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:导航名称"`
	Icon      string    `json:"icon" gorm:"size:255;comment:图标"`
	URL       string    `json:"url" gorm:"size:255;comment:链接地址"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// RolePlugins 角色插件权限关联表
type RolePlugins struct {
	ID       uint `json:"id" gorm:"primaryKey;comment:记录ID"`
	RoleID   uint `json:"role_id" gorm:"uniqueIndex:idx_rpl;not null;comment:角色ID"`
	PluginID uint `json:"plugin_id" gorm:"uniqueIndex:idx_rpl;not null;comment:插件ID"`
}
