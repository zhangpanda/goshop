package model

import "time"

// Config 系统配置表
type Config struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:配置ID"`
	Group     string    `json:"group" gorm:"size:32;index;comment:配置分组"`
	Key       string    `json:"key" gorm:"uniqueIndex;size:64;not null;comment:配置标识"`
	Value     string    `json:"value" gorm:"type:text;comment:配置值"`
	Desc      string    `json:"desc" gorm:"size:255;comment:配置描述"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Region 地区表(省市区)
type Region struct {
	ID       uint     `json:"id" gorm:"primaryKey;comment:地区ID"`
	ParentID uint     `json:"parent_id" gorm:"index;default:0;comment:上级地区ID"`
	Name     string   `json:"name" gorm:"size:64;not null;comment:地区名称"`
	Level    int8     `json:"level" gorm:"comment:层级:1省2市3区"`
	Sort     int      `json:"sort" gorm:"default:0;comment:排序"`
	Children []Region `json:"children,omitempty" gorm:"-"`
}

// Slide 幻灯片/轮播图表
type Slide struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:幻灯片ID"`
	Name      string    `json:"name" gorm:"size:64;comment:幻灯片名称"`
	Images    string    `json:"images" gorm:"type:text;comment:图片JSON数组"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// Navigation 导航菜单表
type Navigation struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:导航ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:导航名称"`
	URL       string    `json:"url" gorm:"size:255;comment:链接地址"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	Type      string    `json:"type" gorm:"size:16;comment:位置:header/footer"`
	CreatedAt time.Time `json:"created_at"`
}

// Link 友情链接表
type Link struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:链接ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:链接名称"`
	URL       string    `json:"url" gorm:"size:255;comment:链接地址"`
	Logo      string    `json:"logo" gorm:"size:255;comment:链接Logo"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// Payment 支付方式配置表
type Payment struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:支付方式ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:支付方式名称"`
	Logo      string    `json:"logo" gorm:"size:255;comment:支付方式图标"`
	Config    string    `json:"config" gorm:"type:text;comment:支付配置JSON"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// SmsLog 短信发送日志表
type SmsLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:日志ID"`
	Phone     string    `json:"phone" gorm:"size:20;index;comment:手机号"`
	Content   string    `json:"content" gorm:"type:text;comment:短信内容"`
	Type      string    `json:"type" gorm:"size:32;comment:短信类型"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0发送中1成功2失败"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailLog 邮件发送日志表
type EmailLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:日志ID"`
	Email     string    `json:"email" gorm:"size:128;index;comment:收件邮箱"`
	Title     string    `json:"title" gorm:"size:255;comment:邮件标题"`
	Content   string    `json:"content" gorm:"type:text;comment:邮件内容"`
	Status    int8      `json:"status" gorm:"default:0;comment:状态:0发送中1成功2失败"`
	CreatedAt time.Time `json:"created_at"`
}

// Attachment 附件/文件表
type Attachment struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:附件ID"`
	CategoryID uint      `json:"category_id" gorm:"index;comment:分类ID"`
	Name       string    `json:"name" gorm:"size:255;comment:文件名"`
	Path       string    `json:"path" gorm:"size:255;not null;comment:文件路径"`
	Size       int64     `json:"size" gorm:"comment:文件大小(字节)"`
	Ext        string    `json:"ext" gorm:"size:16;comment:文件扩展名"`
	CreatedAt  time.Time `json:"created_at"`
}

// AttachmentCategory 附件分类表
type AttachmentCategory struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:分类ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:分类名称"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorLog 错误日志表
type ErrorLog struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:日志ID"`
	Type      string    `json:"type" gorm:"size:32;comment:错误类型"`
	Content   string    `json:"content" gorm:"type:text;comment:错误内容"`
	URL       string    `json:"url" gorm:"size:255;comment:请求URL"`
	IP        string    `json:"ip" gorm:"size:64;comment:请求IP"`
	CreatedAt time.Time `json:"created_at"`
}
