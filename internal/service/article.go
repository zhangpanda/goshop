package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

func CreateArticleCategory(name string, sort int) (*model.ArticleCategory, error) {
	c := model.ArticleCategory{Name: name, Sort: sort, Status: 1}
	return &c, app.Must().DB.Create(&c).Error
}

func GetArticleCategoryList() ([]model.ArticleCategory, error) {
	var list []model.ArticleCategory
	return list, app.Must().DB.Where("status = 1").Order("sort DESC").Find(&list).Error
}

type ArticleReq struct {
	CategoryID uint   `json:"category_id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Cover      string `json:"cover"`
	Author     string `json:"author"`
	Sort       int    `json:"sort"`
}

func CreateArticle(req *ArticleReq) (*model.Article, error) {
	a := model.Article{CategoryID: req.CategoryID, Title: req.Title, Content: req.Content, Cover: req.Cover, Author: req.Author, Sort: req.Sort, Status: 1}
	return &a, app.Must().DB.Create(&a).Error
}

func GetArticleList(categoryID uint, page, pageSize int) ([]model.Article, int64, error) {
	var total int64
	db := app.Must().DB.Model(&model.Article{}).Where("status = 1")
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	db.Count(&total)
	var list []model.Article
	err := db.Preload("Category").Order("sort DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func GetArticleDetail(id uint) (*model.Article, error) {
	var a model.Article
	if err := app.Must().DB.Preload("Category").First(&a, id).Error; err != nil {
		return nil, err
	}
	app.Must().DB.Model(&a).Update("access_count", a.AccessCount+1)
	return &a, nil
}

func DeleteArticle(id uint) error { return app.Must().DB.Delete(&model.Article{}, id).Error }
