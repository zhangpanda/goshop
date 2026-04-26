package model

import "time"

// Plugin 插件表
type Plugin struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:插件ID"`
	Name      string    `json:"name" gorm:"size:64;uniqueIndex;not null;comment:插件标识"`
	Title     string    `json:"title" gorm:"size:128;comment:插件名称"`
	Desc      string    `json:"desc" gorm:"size:255;comment:插件描述"`
	Author    string    `json:"author" gorm:"size:64;comment:作者"`
	Version   string    `json:"version" gorm:"size:16;comment:版本号"`
	Config    string    `json:"config" gorm:"type:text;comment:插件配置JSON"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0未安装1已安装"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PluginCategory 插件分类表
type PluginCategory struct {
	ID   uint   `json:"id" gorm:"primaryKey;comment:分类ID"`
	Name string `json:"name" gorm:"size:64;not null;comment:分类名称"`
	Sort int    `json:"sort" gorm:"default:0;comment:排序"`
}

// Diy DIY页面装修表
type Diy struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:DIY页面ID"`
	Name        string    `json:"name" gorm:"size:64;not null;comment:页面名称"`
	Data        string    `json:"data" gorm:"type:longtext;comment:页面配置JSON"`
	AccessCount int       `json:"access_count" gorm:"default:0;comment:浏览量"`
	Status      int8      `json:"status" gorm:"default:0;comment:状态:0禁用1启用"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CustomView 自定义页面表
type CustomView struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:页面ID"`
	Title     string    `json:"title" gorm:"size:128;not null;comment:页面标题"`
	Content   string    `json:"content" gorm:"type:longtext;comment:页面内容"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ThemeData 主题配置表
type ThemeData struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:主题ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:主题名称"`
	Data      string    `json:"data" gorm:"type:longtext;comment:主题配置JSON"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// FormInput 自定义表单表
type FormInput struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:表单ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:表单名称"`
	Config    string    `json:"config" gorm:"type:longtext;comment:表单配置JSON"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// FormInputData 表单提交数据表
type FormInputData struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	FormID    uint      `json:"form_id" gorm:"index;not null;comment:表单ID"`
	UserID    uint      `json:"user_id" gorm:"index;comment:用户ID"`
	Data      string    `json:"data" gorm:"type:text;comment:提交数据JSON"`
	CreatedAt time.Time `json:"created_at"`
}

// AppHomeNav APP首页导航表
type AppHomeNav struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:导航ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:导航名称"`
	Icon      string    `json:"icon" gorm:"size:255;comment:图标"`
	URL       string    `json:"url" gorm:"size:255;comment:链接地址"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// AppCenterNav APP个人中心导航表
type AppCenterNav struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:导航ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:导航名称"`
	Icon      string    `json:"icon" gorm:"size:255;comment:图标"`
	URL       string    `json:"url" gorm:"size:255;comment:链接地址"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// AppTabbar APP底部导航表
type AppTabbar struct {
	ID         uint   `json:"id" gorm:"primaryKey;comment:导航ID"`
	Name       string `json:"name" gorm:"size:64;not null;comment:导航名称"`
	Icon       string `json:"icon" gorm:"size:255;comment:默认图标"`
	ActiveIcon string `json:"active_icon" gorm:"size:255;comment:选中图标"`
	URL        string `json:"url" gorm:"size:255;comment:链接地址"`
	Sort       int    `json:"sort" gorm:"default:0;comment:排序"`
}

// ShortcutMenu 快捷菜单表
type ShortcutMenu struct {
	ID   uint   `json:"id" gorm:"primaryKey;comment:菜单ID"`
	Name string `json:"name" gorm:"size:64;not null;comment:菜单名称"`
	Icon string `json:"icon" gorm:"size:255;comment:图标"`
	URL  string `json:"url" gorm:"size:255;comment:链接地址"`
	Sort int    `json:"sort" gorm:"default:0;comment:排序"`
}

// Agreement 协议内容表(用户协议/隐私政策)
type Agreement struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:协议ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:协议名称(如用户协议/隐私政策)"`
	Content   string    `json:"content" gorm:"type:longtext;comment:协议内容"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderTraceSource 订单溯源表
type OrderTraceSource struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:记录ID"`
	OrderID   uint      `json:"order_id" gorm:"index;not null;comment:订单ID"`
	UserID    uint      `json:"user_id" gorm:"index;comment:用户ID"`
	Source    string    `json:"source" gorm:"size:64;comment:来源渠道"`
	Params    string    `json:"params" gorm:"type:text;comment:溯源参数JSON"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderCurrency 订单货币信息表
type OrderCurrency struct {
	ID       uint    `json:"id" gorm:"primaryKey;comment:记录ID"`
	OrderID  uint    `json:"order_id" gorm:"index;not null;comment:订单ID"`
	Currency string  `json:"currency" gorm:"size:10;default:CNY;comment:货币代码"`
	Rate     float64 `json:"rate" gorm:"default:1;comment:汇率"`
	Amount   int64   `json:"amount" gorm:"comment:换算后金额(分)"`
}
