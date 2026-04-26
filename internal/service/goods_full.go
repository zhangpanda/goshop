package service

import (
	"fmt"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
)

// GoodsSave 完整保存商品（含规格/参数/相册/分类）
type GoodsSaveReq struct {
	GoodsReq
	SKUs         []SKUReq           `json:"skus"`
	CategoryIDs  []uint             `json:"category_ids"`
	Params       []ParamsConfigItem `json:"params"`
	Photos       []string           `json:"photos"`
	ContentApp   string             `json:"content_app"`
}

func GoodsSave(id uint, req *GoodsSaveReq) (*model.Goods, error) {
	tx := global.DB.Begin()
	if id > 0 {
		tx.Model(&model.Goods{}).Where("id = ?", id).Updates(map[string]interface{}{
			"category_id": req.CategoryID, "title": req.Title, "subtitle": req.Subtitle,
			"main_image": req.MainImage, "images": req.Images, "detail": req.Detail,
		})
		// 重建SKU
		tx.Where("goods_id = ?", id).Delete(&model.GoodsSKU{})
	} else {
		g := model.Goods{CategoryID: req.CategoryID, Title: req.Title, Subtitle: req.Subtitle, MainImage: req.MainImage, Images: req.Images, Detail: req.Detail}
		tx.Create(&g)
		id = g.ID
	}
	for _, s := range req.SKUs {
		tx.Create(&model.GoodsSKU{GoodsID: id, Name: s.Name, Price: s.Price, Stock: s.Stock, Image: s.Image, Specs: s.Specs, Status: 1})
	}
	tx.Commit()
	// 多分类
	if len(req.CategoryIDs) > 0 { SaveGoodsCategoryJoinRecords(id, req.CategoryIDs) }
	// 参数
	if len(req.Params) > 0 { SaveGoodsParams(id, req.Params) }
	// 相册
	if len(req.Photos) > 0 { SaveGoodsPhotos(id, req.Photos) }
	// APP详情
	if req.ContentApp != "" { SaveGoodsContentApp(id, req.ContentApp) }
	var goods model.Goods
	global.DB.Preload("SKUs").Preload("Category").First(&goods, id)
	return &goods, nil
}

// GoodsSaveBaseUpdate 仅更新基础信息
func GoodsSaveBaseUpdate(id uint, updates map[string]interface{}) error {
	return global.DB.Model(&model.Goods{}).Where("id = ?", id).Updates(updates).Error
}

// GoodsData 获取单个商品数据
func GoodsData(id uint) *model.Goods {
	var g model.Goods
	global.DB.Preload("SKUs").Preload("Category").First(&g, id)
	if g.ID == 0 { return nil }
	return &g
}

// GoodsDataEditStatusCheck 编辑状态检查
func GoodsDataEditStatusCheck(id uint) error {
	var g model.Goods
	global.DB.Select("status").First(&g, id)
	if g.Status == 1 { return fmt.Errorf("商品已上架，请先下架再编辑") }
	return nil
}

// GoodsSearchList 商品搜索列表（含关键字+分类+品牌）
func GoodsSearchList(keyword string, categoryID, brandID uint, page, pageSize int) ([]model.Goods, int64) {
	db := global.DB.Model(&model.Goods{}).Where("status = 1")
	if keyword != "" { db = db.Where("title LIKE ?", "%"+keyword+"%") }
	if categoryID > 0 {
		ids := GoodsCategoryItemsIds([]uint{categoryID}, 3)
		db = db.Where("category_id IN ?", ids)
	}
	if brandID > 0 { db = db.Where("brand_id = ?", brandID) }
	var total int64
	db.Count(&total)
	var list []model.Goods
	db.Preload("SKUs").Order("sort DESC, id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list)
	return list, total
}

// AppointGoodsList 指定ID商品列表
func AppointGoodsList(ids []uint) []model.Goods {
	var list []model.Goods
	if len(ids) == 0 { return list }
	global.DB.Where("id IN ? AND status = 1", ids).Preload("SKUs").Find(&list)
	return list
}

// AutoGoodsList 自动商品列表（按条件）
func AutoGoodsList(categoryID uint, orderBy string, limit int) []model.Goods {
	if limit <= 0 { limit = 10 }
	db := global.DB.Where("status = 1")
	if categoryID > 0 { db = db.Where("category_id = ?", categoryID) }
	order := "sort DESC, id DESC"
	switch orderBy {
	case "sales": order = "sales_count DESC"
	case "new": order = "id DESC"
	case "price_asc": order = "id ASC"
	}
	var list []model.Goods
	db.Preload("SKUs").Order(order).Limit(limit).Find(&list)
	return list
}

// CategoryGoodsList 分类下商品列表
func CategoryGoodsList(categoryID uint, page, pageSize int) ([]model.Goods, int64) {
	ids := GoodsCategoryItemsIds([]uint{categoryID}, 3)
	var total int64
	global.DB.Model(&model.Goods{}).Where("category_id IN ? AND status = 1", ids).Count(&total)
	var list []model.Goods
	global.DB.Where("category_id IN ? AND status = 1", ids).Preload("SKUs").
		Order("sort DESC, id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&list)
	return list, total
}

// CategoryGoodsTotal 分类下商品总数
func CategoryGoodsTotal(categoryID uint) int64 {
	ids := GoodsCategoryItemsIds([]uint{categoryID}, 3)
	var c int64
	global.DB.Model(&model.Goods{}).Where("category_id IN ? AND status = 1", ids).Count(&c)
	return c
}

// GoodsUrlCreate 生成商品URL
func GoodsUrlCreate(id uint) string { return fmt.Sprintf("/products/%d", id) }

// GoodsQrcode 商品二维码
func GoodsQrcode(id uint) string { return GenerateQRCodeURL(GoodsUrlCreate(id)) }

// GoodsImagesCoverHandle 封面图处理
func GoodsImagesCoverHandle(goods *model.Goods) string {
	if goods.MainImage != "" { return goods.MainImage }
	if len(goods.SKUs) > 0 && goods.SKUs[0].Image != "" { return goods.SKUs[0].Image }
	return ""
}

// GoodsSpecificationsData 获取商品规格数据
func GoodsSpecificationsData(goodsID uint) ([]model.GoodsSKU, error) {
	var list []model.GoodsSKU
	return list, global.DB.Where("goods_id = ? AND status = 1", goodsID).Find(&list).Error
}

// GoodsSpecificationsActual 获取实际规格组合
func GoodsSpecificationsActual(goodsID uint) ([]model.GoodsSpecBase, error) {
	return GetGoodsSpecBase(goodsID)
}

// GoodsSpecificationsInsert 保存规格
func GoodsSpecificationsInsert(goodsID uint, specs []SpecBaseReq) error {
	return SaveGoodsSpecBase(goodsID, specs)
}

// GoodsSpecDefaultName 默认规格名称
func GoodsSpecDefaultName() string { return "默认" }

// GoodsSpecBaseFields 规格基础字段
func GoodsSpecBaseFields() []string {
	return []string{"price", "original_price", "inventory", "weight", "volume", "coding", "barcode"}
}

// GoodsSpecType 获取规格类型
func GoodsSpecType(goodsID uint) []model.SpecType {
	var list []model.SpecType
	global.DB.Preload("Values").Where("template_id IN (?)",
		global.DB.Model(&model.SpecTemplate{}).Select("id")).Find(&list)
	return list
}

// GoodsSpecOperateData 规格操作数据
func GoodsSpecOperateData(goodsID uint) map[string]interface{} {
	specs, _ := GoodsSpecificationsData(goodsID)
	bases, _ := GoodsSpecificationsActual(goodsID)
	return map[string]interface{}{"skus": specs, "spec_base": bases}
}

// GoodsParametersData 商品参数数据
func GoodsParametersData(goodsID uint) ([]model.GoodsParams, error) { return GetGoodsParams(goodsID) }

// GoodsParamsInsert 保存参数
func GoodsParamsInsert(goodsID uint, params []ParamsConfigItem) error { return SaveGoodsParams(goodsID, params) }

// GoodsParamsOperateData 参数操作数据
func GoodsParamsOperateData(goodsID uint) map[string]interface{} {
	params, _ := GetGoodsParams(goodsID)
	templates, _ := GetParamsTemplateList()
	return map[string]interface{}{"params": params, "templates": templates}
}

// GoodsPhotoData 商品相册数据
func GoodsPhotoData(goodsID uint) ([]model.GoodsPhoto, error) { return GetGoodsPhotos(goodsID) }

// GoodsPhotoInsert 保存相册
func GoodsPhotoInsert(goodsID uint, images []string) error { return SaveGoodsPhotos(goodsID, images) }

// GoodsCategoryInsert 保存商品分类关联
func GoodsCategoryInsert(goodsID uint, catIDs []uint) { SaveGoodsCategoryJoinRecords(goodsID, catIDs) }

// GoodsContentAppInsert 保存APP详情
func GoodsContentAppInsert(goodsID uint, content string) error { return SaveGoodsContentApp(goodsID, content) }

// GoodsEditParameters 编辑页参数数据
func GoodsEditParameters(goodsID uint) map[string]interface{} { return GoodsParamsOperateData(goodsID) }

// GoodsEditSpecifications 编辑页规格数据
func GoodsEditSpecifications(goodsID uint) map[string]interface{} { return GoodsSpecOperateData(goodsID) }

// GoodsBuyButtonList 购买按钮列表
func GoodsBuyButtonList(goods *model.Goods) []map[string]string {
	if goods.Status != 1 { return nil }
	btns := []map[string]string{{"type": "buy", "name": "立即购买"}, {"type": "cart", "name": "加入购物车"}}
	return btns
}

// GoodsSalesModelType 销售模式类型
func GoodsSalesModelType(goods *model.Goods) string {
	siteType := GetConfig("common_site_type")
	switch siteType {
	case "2": return "自提"
	case "3": return "虚拟"
	case "4": return "展示"
	default: return "快递"
	}
}

// GoodsDetailMiddleTabsNavList 详情页中间导航
func GoodsDetailMiddleTabsNavList(goodsID uint) []map[string]interface{} {
	tabs := []map[string]interface{}{
		{"type": "detail", "name": "详情", "active": true},
	}
	if GoodsCommentsTotal(goodsID) > 0 {
		tabs = append(tabs, map[string]interface{}{"type": "comments", "name": fmt.Sprintf("评价(%d)", GoodsCommentsTotal(goodsID))})
	}
	return tabs
}

// GoodsDetailSeeingYouData 看了又看
func GoodsDetailSeeingYouData(goodsID uint, limit int) []model.Goods {
	if limit <= 0 { limit = 6 }
	var g model.Goods
	global.DB.Select("category_id").First(&g, goodsID)
	var list []model.Goods
	global.DB.Where("category_id = ? AND id != ? AND status = 1", g.CategoryID, goodsID).
		Preload("SKUs").Order("sales_count DESC").Limit(limit).Find(&list)
	return list
}

// GoodsDetailGuessYouLikeData 猜你喜欢（详情页）
func GoodsDetailGuessYouLikeData(goodsID uint, limit int) []model.Goods {
	return GoodsDetailSeeingYouData(goodsID, limit)
}

// GoodsListCategoryGroupList 商品列表按分类分组
func GoodsListCategoryGroupList() []HomeFloor { return HomeFloorList(8) }

// GoodsAppData APP端商品数据
func GoodsAppData(goodsID uint) map[string]interface{} {
	goods := GoodsData(goodsID)
	if goods == nil { return nil }
	return map[string]interface{}{
		"goods": goods, "score": GoodsScore(goodsID),
		"photos": func() []model.GoodsPhoto { l, _ := GoodsPhotoData(goodsID); return l }(),
		"content_app": GetGoodsContentApp(goodsID),
	}
}

// GoodsBaseTemplate 商品基础模板数据
func GoodsBaseTemplate() map[string]interface{} {
	cats, _ := GetCategoryTree()
	brands, _ := GetBrandList()
	return map[string]interface{}{"categories": cats, "brands": brands}
}

// GoodsBaseFieldsRequiredConfigData 基础字段必填配置
func GoodsBaseFieldsRequiredConfigData() map[string]bool {
	return map[string]bool{"title": true, "category_id": true, "price": true}
}

// IsGoodsSiteTypeConsistent 商品站点类型一致性检查
func IsGoodsSiteTypeConsistent(goodsID uint) bool { return true }

// UserCartGoodsCountData 用户购物车商品数
func UserCartGoodsCountData(userID uint) int64 { return GoodsCartTotal(userID) }

// UserFavorGoodsCountData 用户收藏商品数
func UserFavorGoodsCountData(userID uint) int64 { return GoodsFavorTotal(userID) }

// GetFormGoodsSpecificationsParams 表单规格参数
func GetFormGoodsSpecificationsParams(goodsID uint) map[string]interface{} { return GoodsSpecOperateData(goodsID) }

// GetFormGoodsSpecificationsBaseParams 表单规格基础参数
func GetFormGoodsSpecificationsBaseParams(goodsID uint) ([]model.GoodsSpecBase, error) { return GoodsSpecificationsActual(goodsID) }

// GetFormGoodsPhotoParams 表单相册参数
func GetFormGoodsPhotoParams(goodsID uint) ([]model.GoodsPhoto, error) { return GoodsPhotoData(goodsID) }

// GetFormGoodsContentAppParams 表单APP详情参数
func GetFormGoodsContentAppParams(goodsID uint) string { return GetGoodsContentApp(goodsID) }

// GoodsBaseForbidOperateData 禁止操作数据
func GoodsBaseForbidOperateData(goodsID uint) map[string]bool { return map[string]bool{} }

// GoodsBuyLeftNavList 购买左侧导航
func GoodsBuyLeftNavList() []map[string]string { return nil }

// GoodsSpecificationsConcise 规格简洁数据
func GoodsSpecificationsConcise(goodsID uint) []string {
	var names []string
	global.DB.Model(&model.GoodsSKU{}).Where("goods_id = ? AND status = 1", goodsID).Pluck("name", &names)
	return names
}

// GoodsSpecificationsExtends 规格扩展数据
func GoodsSpecificationsExtends(goodsID uint) map[string]interface{} { return GoodsSpecOperateData(goodsID) }
