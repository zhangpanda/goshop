package model

import "time"

// ArticleCategory 文章分类表
type ArticleCategory struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:分类ID"`
	Name      string    `json:"name" gorm:"size:64;not null;comment:分类名称"`
	Sort      int       `json:"sort" gorm:"default:0;comment:排序"`
	Status    int8      `json:"status" gorm:"default:1;comment:状态:0禁用1启用"`
	CreatedAt time.Time `json:"created_at"`
}

// Article 文章表
type Article struct {
	ID          uint             `json:"id" gorm:"primaryKey;comment:文章ID"`
	CategoryID  uint             `json:"category_id" gorm:"index;comment:分类ID"`
	Title       string           `json:"title" gorm:"size:255;not null;comment:文章标题"`
	Content     string           `json:"content" gorm:"type:longtext;comment:文章内容"`
	Cover       string           `json:"cover" gorm:"size:255;comment:封面图"`
	Author      string           `json:"author" gorm:"size:64;comment:作者"`
	AccessCount int              `json:"access_count" gorm:"default:0;comment:浏览量"`
	Sort        int              `json:"sort" gorm:"default:0;comment:排序"`
	Status      int8             `json:"status" gorm:"default:1;comment:状态:0草稿1发布"`
	Category    *ArticleCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
