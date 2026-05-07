package service

import (
	"errors"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// ============================================================
// 通用 CRUD 辅助：StatusUpdate / Delete / Total / Save(编辑)
// 供管理端与部分兼容接口复用（非「实现 ShopXO 全部 PHP 模块」）
// ============================================================

// ---- 通用状态更新 ----
func statusUpdate(table string, id uint, field string, val interface{}) error {
	return global.DB.Table(table).Where("id = ?", id).Update(field, val).Error
}
func softDelete(table string, id uint) error {
	return global.DB.Table(table).Where("id = ?", id).Delete(nil).Error
}
func totalCount(m interface{}, where ...interface{}) int64 {
	var c int64
	db := global.DB.Model(m)
	if len(where) >= 2 {
		db = db.Where(where[0], where[1:]...)
	}
	db.Count(&c)
	return c
}

// ==================== Goods ====================
func GoodsStatusUpdate(id uint, status int8) error {
	return statusUpdate("goods", id, "status", status)
}
func GoodsDeleteFull(id uint) error {
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsSKU{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsSpecBase{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsPhoto{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsParams{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsCategoryJoin{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsContentApp{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Goods{}, id).Error
	})
}
func GoodsTotal(status *int8) int64 {
	db := global.DB.Model(&model.Goods{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var c int64
	db.Count(&c)
	return c
}
func GoodsListHandle(list []model.Goods) []model.Goods {
	for i := range list {
		if list[i].SKUs == nil {
			global.DB.Where("goods_id = ?", list[i].ID).Find(&list[i].SKUs)
		}
	}
	return list
}

// ==================== GoodsCategory ====================
func GoodsCategoryStatusUpdate(id uint, status int8) error {
	return statusUpdate("categories", id, "status", status)
}
func GoodsCategoryDelete(id uint) error {
	var count int64
	global.DB.Model(&model.Category{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errHasChildren
	}
	return global.DB.Delete(&model.Category{}, id).Error
}
func GoodsCategoryItemsIds(ids []uint, deep int) []uint {
	result := make([]uint, 0, len(ids))
	result = append(result, ids...)
	if deep <= 0 {
		deep = 3
	}
	for d := 0; d < deep; d++ {
		var childIDs []uint
		global.DB.Model(&model.Category{}).Where("parent_id IN ? AND status = 1", result).Pluck("id", &childIDs)
		if len(childIDs) == 0 {
			break
		}
		result = append(result, childIDs...)
	}
	return result
}
func GoodsCategoryLevel() int {
	v := GetConfig("goods_category_level")
	if v == "" {
		return 3
	}
	n := 3
	for _, c := range v {
		n = int(c - '0')
		break
	}
	if n < 1 || n > 3 {
		n = 3
	}
	return n
}
func GoodsCategoryName(id uint) string {
	var c model.Category
	global.DB.Select("name").First(&c, id)
	return c.Name
}
func GoodsCategoryNames(ids []uint) map[uint]string {
	var cats []model.Category
	global.DB.Where("id IN ?", ids).Select("id, name").Find(&cats)
	m := make(map[uint]string, len(cats))
	for _, c := range cats {
		m[c.ID] = c.Name
	}
	return m
}
func GoodsCategoryParentIds(id uint) []uint {
	var ids []uint
	for id > 0 {
		var c model.Category
		global.DB.Select("id, parent_id").First(&c, id)
		if c.ID == 0 {
			break
		}
		ids = append([]uint{c.ID}, ids...)
		id = c.ParentID
	}
	return ids
}
func GoodsCategoryAll() []model.Category {
	var cats []model.Category
	global.DB.Where("status = 1").Order("sort DESC, id ASC").Find(&cats)
	return buildCategoryTree(cats, 0)
}
func buildCategoryTree(all []model.Category, pid uint) []model.Category {
	var result []model.Category
	for _, c := range all {
		if c.ParentID == pid {
			c.Children = buildCategoryTree(all, c.ID)
			result = append(result, c)
		}
	}
	return result
}

// ==================== GoodsComments ====================
func GoodsCommentsStatusUpdate(id uint, status int8) error {
	return statusUpdate("reviews", id, "status", status)
}
func GoodsCommentsDelete(id uint) error { return global.DB.Delete(&model.Review{}, id).Error }

/**
 * GoodsCommentsDeleteForUser 仅允许删除当前用户的评价。
 */
func GoodsCommentsDeleteForUser(id, userID uint) error {
	res := global.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Review{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("评价不存在或无权删除")
	}
	return nil
}
func GoodsCommentsTotal(goodsID uint) int64 {
	return totalCount(&model.Review{}, "goods_id = ?", goodsID)
}
func GoodsFirstSeveralComments(goodsID uint, n int) []model.Review {
	var list []model.Review
	global.DB.Where("goods_id = ?", goodsID).Preload("User").Order("id DESC").Limit(n).Find(&list)
	return list
}

// ==================== GoodsFavor ====================
func GoodsFavorTotal(userID uint) int64 { return totalCount(&model.Favorite{}, "user_id = ?", userID) }
func GoodsFavorDelete(ids []uint, userID uint) error {
	return global.DB.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.Favorite{}).Error
}

// ==================== GoodsBrowse ====================
func GoodsBrowseDelete(ids []uint, userID uint) error {
	return global.DB.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.BrowseHistory{}).Error
}

// ==================== GoodsCart ====================
func GoodsCartTotal(userID uint) int64 { return totalCount(&model.Cart{}, "user_id = ?", userID) }
func GoodsCartStock(cartID uint) (int, error) {
	var cart model.Cart
	global.DB.Preload("SKU").First(&cart, cartID)
	if cart.SKU == nil {
		return 0, errNotFound
	}
	return cart.SKU.Stock, nil
}

// ==================== Brand ====================
func BrandStatusUpdate(id uint, status int8) error {
	return statusUpdate("brands", id, "status", status)
}
func BrandDeleteFull(id uint) error { return global.DB.Delete(&model.Brand{}, id).Error }
func BrandTotal() int64             { return totalCount(&model.Brand{}, "status = ?", 1) }

// ==================== BrandCategory ====================
func BrandCategoryDelete(id uint) error { return global.DB.Delete(&model.BrandCategory{}, id).Error }

// ==================== Article ====================
func ArticleStatusUpdate(id uint, status int8) error {
	return statusUpdate("articles", id, "status", status)
}
func ArticleDeleteFull(id uint) error { return global.DB.Delete(&model.Article{}, id).Error }
func ArticleTotal(catID uint) int64 {
	db := global.DB.Model(&model.Article{}).Where("status = 1")
	if catID > 0 {
		db = db.Where("category_id = ?", catID)
	}
	var c int64
	db.Count(&c)
	return c
}

// ==================== ArticleCategory ====================
func ArticleCategoryStatusUpdate(id uint, status int8) error {
	return statusUpdate("article_categories", id, "status", status)
}
func ArticleCategoryDelete(id uint) error {
	var count int64
	global.DB.Model(&model.Article{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return errHasChildren
	}
	return global.DB.Delete(&model.ArticleCategory{}, id).Error
}

// ==================== Express ====================
func ExpressStatusUpdate(id uint, status int8) error {
	return statusUpdate("expresses", id, "status", status)
}
func ExpressDelete(id uint) error { return global.DB.Delete(&model.Express{}, id).Error }

// ==================== Slide ====================
func SlideStatusUpdate(id uint, status int8) error {
	return statusUpdate("slides", id, "status", status)
}
func SlideDelete(id uint) error { return global.DB.Delete(&model.Slide{}, id).Error }

// ==================== Navigation ====================
func NavStatusUpdate(id uint, status int8) error {
	return statusUpdate("navigations", id, "status", status)
}
func NavDelete(id uint) error { return global.DB.Delete(&model.Navigation{}, id).Error }
func NavSave(n *model.Navigation) error {
	if n.ID > 0 {
		return global.DB.Save(n).Error
	}
	return global.DB.Create(n).Error
}
func BottomNavigationData() []model.Navigation {
	var l []model.Navigation
	global.DB.Where("type = 'footer' AND status = 1").Order("sort DESC").Find(&l)
	return l
}
func LevelOneNav() []model.Navigation {
	var l []model.Navigation
	global.DB.Where("type = 'header' AND status = 1").Order("sort DESC").Find(&l)
	return l
}
func UserCenterLeftList() []model.Navigation {
	var l []model.Navigation
	global.DB.Where("type = 'user_center' AND status = 1").Order("sort DESC").Find(&l)
	return l
}

// ==================== Link ====================
func LinkStatusUpdate(id uint, status int8) error { return statusUpdate("links", id, "status", status) }
func LinkDelete(id uint) error                    { return global.DB.Delete(&model.Link{}, id).Error }

// ==================== ScreeningPrice ====================
func ScreeningPriceSave(s *model.ScreeningPrice) error {
	if s.ID > 0 {
		return global.DB.Save(s).Error
	}
	return global.DB.Create(s).Error
}
func ScreeningPriceDelete(id uint) error { return global.DB.Delete(&model.ScreeningPrice{}, id).Error }

// ==================== Region ====================
func RegionSave(r *model.Region) error {
	if r.ID > 0 {
		return global.DB.Save(r).Error
	}
	return global.DB.Create(r).Error
}
func RegionDelete(id uint) error {
	var count int64
	global.DB.Model(&model.Region{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errHasChildren
	}
	return global.DB.Delete(&model.Region{}, id).Error
}
func RegionStatusUpdate(id uint, status int8) error {
	return statusUpdate("regions", id, "status", status)
}
func RegionAll() []model.Region {
	var list []model.Region
	global.DB.Order("sort, id").Find(&list)
	return list
}
func RegionCodeData(id uint) *model.Region {
	var r model.Region
	global.DB.First(&r, id)
	return &r
}
func RegionName(id uint) string {
	var r model.Region
	global.DB.Select("name").First(&r, id)
	return r.Name
}
func RegionNodeSon(pid uint) []model.Region {
	var list []model.Region
	global.DB.Where("parent_id = ?", pid).Order("sort, id").Find(&list)
	return list
}
func RegionItemsIds(pid uint) []uint {
	var ids []uint
	global.DB.Model(&model.Region{}).Where("parent_id = ?", pid).Pluck("id", &ids)
	return ids
}

// ==================== CustomView ====================
func CustomViewStatusUpdate(id uint, status int8) error {
	return statusUpdate("custom_views", id, "status", status)
}
func CustomViewDelete(id uint) error { return global.DB.Delete(&model.CustomView{}, id).Error }

// ==================== Design ====================
func DesignStatusUpdate(id uint, status int8) error {
	return statusUpdate("designs", id, "status", status)
}
func DesignDelete(id uint) error { return global.DB.Delete(&model.Design{}, id).Error }
func DesignAccessCountInc(id uint) {
	global.DB.Model(&model.Design{}).Where("id = ?", id).Update("status", gorm.Expr("COALESCE(status,0)"))
}

// ==================== Diy ====================
func DiyStatusUpdate(id uint, status int8) error { return statusUpdate("diys", id, "status", status) }
func DiyAccessCountInc(id uint) {
	global.DB.Model(&model.Diy{}).Where("id = ?", id).Update("access_count", gorm.Expr("access_count + 1"))
}

// ==================== FormInput ====================
func FormInputStatusUpdate(id uint, status int8) error {
	return statusUpdate("form_inputs", id, "status", status)
}
func FormInputDelete(id uint) error {
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Where("form_id = ?", id).Delete(&model.FormInputData{}).Error; err != nil {
			return err
		}
		if err := tx.Where("form_id = ?", id).Delete(&model.FormTableUserFields{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.FormInput{}, id).Error
	})
}

// ==================== AppHomeNav ====================
func AppHomeNavStatusUpdate(id uint, status int8) error {
	return statusUpdate("app_home_navs", id, "status", status)
}
func AppHomeNavDelete(id uint) error { return global.DB.Delete(&model.AppHomeNav{}, id).Error }

// ==================== AppCenterNav ====================
func AppCenterNavStatusUpdate(id uint, status int8) error {
	return statusUpdate("app_center_navs", id, "status", status)
}
func AppCenterNavDelete(id uint) error { return global.DB.Delete(&model.AppCenterNav{}, id).Error }

// ==================== ShortcutMenu ====================
func ShortcutMenuDelete(id uint) error { return global.DB.Delete(&model.ShortcutMenu{}, id).Error }

// ==================== QuickNav ====================
func QuickNavStatusUpdate(id uint, status int8) error {
	return statusUpdate("quick_navs", id, "status", status)
}
func QuickNavDelete(id uint) error { return global.DB.Delete(&model.QuickNav{}, id).Error }

// ==================== Attachment ====================
func AttachmentDelete(id uint) error { return global.DB.Delete(&model.Attachment{}, id).Error }
func AttachmentSave(a *model.Attachment) error {
	if a.ID > 0 {
		return global.DB.Save(a).Error
	}
	return global.DB.Create(a).Error
}

// ==================== AttachmentCategory ====================
func AttachmentCategoryStatusUpdate(id uint, status int8) error {
	return statusUpdate("attachment_categories", id, "status", status)
}
func AttachmentCategoryDelete(id uint) error {
	return global.DB.Delete(&model.AttachmentCategory{}, id).Error
}

// ==================== ErrorLog ====================
func ErrorLogDelete(ids []uint) error {
	return global.DB.Where("id IN ?", ids).Delete(&model.ErrorLog{}).Error
}
func ErrorLogAllDelete() error { return global.DB.Where("1=1").Delete(&model.ErrorLog{}).Error }

// ==================== PluginsCategory ====================
func PluginsCategorySave(name string, sort int) error {
	return global.DB.Create(&model.PluginCategory{Name: name, Sort: sort}).Error
}
func PluginsCategoryDelete(id uint) error { return global.DB.Delete(&model.PluginCategory{}, id).Error }
func PluginsCategoryList() []model.PluginCategory {
	var l []model.PluginCategory
	global.DB.Order("sort DESC").Find(&l)
	return l
}

// ==================== Admin ====================
func AdminDetail(id uint) (*model.Admin, error) {
	var a model.Admin
	if err := global.DB.Preload("Role").First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}
func AdminSave(id uint, nickname string, roleID uint) error {
	return global.DB.Model(&model.Admin{}).Where("id = ?", id).Updates(map[string]interface{}{"nickname": nickname, "role_id": roleID}).Error
}
func AdminDelete(id uint) error { return global.DB.Delete(&model.Admin{}, id).Error }

// ==================== Power ====================
func PowerStatusUpdate(id uint, status int8) error {
	return statusUpdate("powers", id, "status", status)
}
func PowerSave(p *model.Power) error {
	if p.ID > 0 {
		return global.DB.Save(p).Error
	}
	return global.DB.Create(p).Error
}

// ==================== Message ====================
func MessageTotal(userID uint) int64 { return totalCount(&model.Message{}, "user_id = ?", userID) }
func MessageDelete(ids []uint, userID uint) error {
	return global.DB.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.Message{}).Error
}

// ==================== 通用错误 ====================
var (
	errHasChildren = errMsg("请先删除子数据")
	errNotFound    = errMsg("数据不存在")
)

type errMsg string

func (e errMsg) Error() string { return string(e) }
