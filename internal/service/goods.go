package service

import (
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// ---- 分类 ----

type CategoryReq struct {
	ParentID uint   `json:"parent_id"`
	Name     string `json:"name" binding:"required"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
}

func CreateCategory(req *CategoryReq) (*model.Category, error) {
	cat := model.Category{
		ParentID: req.ParentID,
		Name:     req.Name,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   1,
	}
	if err := global.DB.Create(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func GetCategoryTree() ([]model.Category, error) {
	var cats []model.Category
	err := global.DB.Where("parent_id = 0 AND status = 1").
		Order("sort DESC").Find(&cats).Error
	if err != nil {
		return nil, err
	}
	for i := range cats {
		global.DB.Where("parent_id = ? AND status = 1", cats[i].ID).Order("sort DESC").Find(&cats[i].Children)
	}
	return cats, nil
}

// ---- 商品 ----

type GoodsReq struct {
	CategoryID uint   `json:"category_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Subtitle   string `json:"subtitle"`
	MainImage  string `json:"main_image"`
	Images     string `json:"images"`
	Detail     string `json:"detail"`
}

type SKUReq struct {
	Name   string `json:"name" binding:"required"`
	Price  int64  `json:"price" binding:"required,min=1"`
	Stock  int    `json:"stock"`
	Image  string `json:"image"`
	Specs  string `json:"specs"`
	Coding string `json:"coding"`
}

type CreateGoodsReq struct {
	GoodsReq
	SKUs []SKUReq `json:"skus" binding:"required,min=1"`
}

func CreateGoods(req *CreateGoodsReq) (*model.Goods, error) {
	goods := model.Goods{
		CategoryID: req.CategoryID,
		Title:      req.Title,
		Subtitle:   req.Subtitle,
		MainImage:  req.MainImage,
		Images:     req.Images,
		Detail:     req.Detail,
	}

	tx := global.DB.Begin()
	if err := tx.Create(&goods).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, s := range req.SKUs {
		sku := model.GoodsSKU{
			GoodsID: goods.ID,
			Name:    s.Name,
			Price:   s.Price,
			Stock:   s.Stock,
			Image:   s.Image,
			Specs:   s.Specs,
			Coding:  s.Coding,
			Status:  1,
		}
		if err := tx.Create(&sku).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	// 重新查询带关联
	global.DB.Preload("SKUs").Preload("Category").First(&goods, goods.ID)
	return &goods, nil
}

type GoodsListReq struct {
	CategoryID uint   `form:"category_id"`
	Keyword    string `form:"keyword"`
	Status     *int8  `form:"status"`
	BrandID    uint   `form:"brand_id"`
	MinPrice   int64  `form:"min_price"`
	MaxPrice   int64  `form:"max_price"`
	SpecValues string `form:"spec_values"` // 逗号分隔，如 "红色,256GB"
	ParamName  string `form:"param_name"`
	ParamValue string `form:"param_value"`
	Region     string `form:"region"`   // 产地
	OrderBy    string `form:"order_by"` // price_asc, price_desc, sales, new
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
}

type GoodsListResp struct {
	Total int64         `json:"total"`
	List  []model.Goods `json:"list"`
}

func GetGoodsList(req *GoodsListReq) (*GoodsListResp, error) {
	db := global.DB.Model(&model.Goods{})

	if req.CategoryID > 0 {
		db = db.Where("category_id = ?", req.CategoryID)
	}
	if req.Keyword != "" {
		// 先尝试条码/编码精确匹配
		var barcodeGoodsIDs []uint
		global.DB.Model(&model.GoodsSpecBase{}).Where("barcode = ? OR coding = ?", req.Keyword, req.Keyword).
			Pluck("goods_id", &barcodeGoodsIDs)
		if len(barcodeGoodsIDs) > 0 {
			db = db.Where("id IN ?", barcodeGoodsIDs)
		} else {
			// 标题+SEO字段模糊搜索
			db = db.Where("title LIKE ? OR subtitle LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
		}
		// 记录搜索热词
		go func() { global.DB.Create(&model.SearchHistory{Keyword: req.Keyword}) }()
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if req.BrandID > 0 {
		db = db.Where("brand_id = ?", req.BrandID)
	}
	if req.MinPrice > 0 || req.MaxPrice > 0 {
		subQ := global.DB.Model(&model.GoodsSKU{}).Select("goods_id")
		if req.MinPrice > 0 {
			subQ = subQ.Where("price >= ?", req.MinPrice)
		}
		if req.MaxPrice > 0 {
			subQ = subQ.Where("price <= ?", req.MaxPrice)
		}
		db = db.Where("id IN (?)", subQ)
	}
	if req.SpecValues != "" {
		db = db.Where("id IN (?)", global.DB.Model(&model.GoodsSpecBase{}).Select("goods_id").Where("spec_values LIKE ?", "%"+req.SpecValues+"%"))
	}
	if req.ParamName != "" {
		pq := global.DB.Model(&model.GoodsParams{}).Select("goods_id").Where("name = ?", req.ParamName)
		if req.ParamValue != "" {
			pq = pq.Where("value = ?", req.ParamValue)
		}
		db = db.Where("id IN (?)", pq)
	}
	if req.Region != "" {
		db = db.Where("id IN (?)", global.DB.Model(&model.GoodsParams{}).Select("goods_id").Where("name = '产地' AND value LIKE ?", "%"+req.Region+"%"))
	}

	var total int64
	db.Count(&total)

	var list []model.Goods
	offset := (req.Page - 1) * req.PageSize
	orderClause := "sort DESC, id DESC"
	switch req.OrderBy {
	case "price_asc":
		orderClause = "id ASC" // 简化，实际应join SKU
	case "price_desc":
		orderClause = "id DESC"
	case "sales":
		orderClause = "sales_count DESC, id DESC"
	case "new":
		orderClause = "id DESC"
	}
	err := db.Preload("SKUs").Preload("Category").
		Order(orderClause).
		Offset(offset).Limit(req.PageSize).
		Find(&list).Error

	return &GoodsListResp{Total: total, List: list}, err
}

func GetGoodsDetail(id uint) (*model.Goods, error) {
	var goods model.Goods
	err := global.DB.Preload("SKUs").Preload("Category").First(&goods, id).Error
	if err != nil {
		return nil, err
	}
	return &goods, nil
}
