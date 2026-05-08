package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 通用状态更新 ==========
func statusUpdate(c *gin.Context, m interface{}) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status int8 `json:"status"`
	}
	c.ShouldBindJSON(&req)
	app.Must().DB.Model(m).Where("id = ?", id).Update("status", req.Status)
	response.OK(c, nil)
}

func genericDelete(c *gin.Context, m interface{}) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	app.Must().DB.Delete(m, id)
	response.OK(c, nil)
}

func genericDetail(c *gin.Context, m interface{}) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := app.Must().DB.First(m, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "不存在")
		return
	}
	response.OK(c, m)
}

func genericUpdate(c *gin.Context, m interface{}) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := c.ShouldBindJSON(m); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	app.Must().DB.Model(m).Where("id = ?", id).Updates(m)
	response.OK(c, nil)
}

// ========== 管理员 ==========
func AdminDetailHandler(c *gin.Context) { var m model.Admin; genericDetail(c, &m) }
func AdminDeleteHandler(c *gin.Context) { genericDelete(c, &model.Admin{}) }

// ========== 角色 ==========
func RoleDetailHandler(c *gin.Context)       { var m model.Role; genericDetail(c, &m) }
func RoleStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Role{}) }

// ========== 商品分类 ==========
func CategoryStatusUpdate(c *gin.Context) {
	statusUpdate(c, &model.Category{})
	service.InvalidateCategoryCache()
}

// ========== 商品评论 ==========
func ReviewDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Review{}) }
func ReviewStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Review{}) }

// ========== 订单 ==========
func AdminCancelOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	app.Must().DB.Model(&model.Order{}).Where("id = ? AND status IN (0,1)", id).Update("status", model.OrderStatusCancelled)
	response.OK(c, nil)
}
func AdminConfirmReceive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	app.Must().DB.Model(&model.Order{}).Where("id = ? AND status = 2", id).Update("status", model.OrderStatusCompleted)
	response.OK(c, nil)
}
func AdminDeleteOrder(c *gin.Context) { genericDelete(c, &model.Order{}) }

// ========== 售后 ==========
func AdminAftersaleDelete(c *gin.Context) { genericDelete(c, &model.OrderAftersale{}) }
func AdminAftersaleCancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	app.Must().DB.Model(&model.OrderAftersale{}).Where("id = ?", id).Update("status", model.AftersaleStatusCancelled)
	response.OK(c, nil)
}

// ========== 品牌 ==========
func BrandUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.Brand{}) }
func BrandDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Brand{}) }
func BrandStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Brand{}) }
func BrandDetailHandler(c *gin.Context)       { var m model.Brand; genericDetail(c, &m) }

// ========== 文章 ==========
func ArticleUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.Article{}) }
func ArticleDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Article{}) }
func ArticleStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Article{}) }
func ArticleDetailHandler(c *gin.Context) {
	var m model.Article
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	app.Must().DB.Preload("Category").First(&m, id)
	response.OK(c, m)
}

// ========== 文章分类 ==========
func ArticleCategoryDeleteHandler(c *gin.Context) { genericDelete(c, &model.ArticleCategory{}) }

// ========== 幻灯片 ==========
func SlideUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.Slide{}) }
func SlideDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Slide{}) }
func SlideStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Slide{}) }
func SlideDetailHandler(c *gin.Context)       { var m model.Slide; genericDetail(c, &m) }

// ========== 导航 ==========
func NavigationUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.Navigation{}) }
func NavigationDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Navigation{}) }
func NavigationStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Navigation{}) }

// ========== 友情链接 ==========
func LinkUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.Link{}) }
func LinkStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Link{}) }
func LinkDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Link{}) }
func LinkDetailHandler(c *gin.Context)       { var m model.Link; genericDetail(c, &m) }

// ========== 快递 ==========
func ExpressUpdateHandler(c *gin.Context) { genericUpdate(c, &model.Express{}) }
func ExpressDeleteHandler(c *gin.Context) { genericDelete(c, &model.Express{}) }

// ========== 支付方式 ==========
func PaymentUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.Payment{}) }
func PaymentStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.Payment{}) }
func PaymentDeleteHandler(c *gin.Context)       { genericDelete(c, &model.Payment{}) }
func PaymentDetailHandler(c *gin.Context)       { var m model.Payment; genericDetail(c, &m) }

// ========== 自定义页面 ==========
func CustomViewUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.CustomView{}) }
func CustomViewDeleteHandler(c *gin.Context)       { genericDelete(c, &model.CustomView{}) }
func CustomViewStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.CustomView{}) }

// ========== 快捷导航 ==========
func QuickNavUpdateHandler(c *gin.Context)       { genericUpdate(c, &model.QuickNav{}) }
func QuickNavDeleteHandler(c *gin.Context)       { genericDelete(c, &model.QuickNav{}) }
func QuickNavStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.QuickNav{}) }

// ========== 附件 ==========
func AttachmentDeleteHandler(c *gin.Context) { genericDelete(c, &model.Attachment{}) }

// ========== 附件分类 ==========
func AttachmentCategoryDeleteHandler(c *gin.Context) { genericDelete(c, &model.AttachmentCategory{}) }
func AttachmentCategoryStatusUpdateHandler(c *gin.Context) {
	statusUpdate(c, &model.AttachmentCategory{})
}

// ========== 搜索记录 ==========
func SearchHistoryDeleteHandler(c *gin.Context) { genericDelete(c, &model.SearchHistory{}) }
func SearchHistoryAllDeleteHandler(c *gin.Context) {
	app.Must().DB.Where("1=1").Delete(&model.SearchHistory{})
	response.OK(c, nil)
}

// ========== 错误日志 ==========
func ErrorLogDeleteHandler(c *gin.Context) { genericDelete(c, &model.ErrorLog{}) }
func ErrorLogAllDeleteHandler(c *gin.Context) {
	app.Must().DB.Where("1=1").Delete(&model.ErrorLog{})
	response.OK(c, nil)
}

// ========== 消息 ==========
func MessageDeleteHandler(c *gin.Context) { genericDelete(c, &model.Message{}) }

// ========== 支付日志 ==========
func PayLogDetailHandler(c *gin.Context) { var m model.PayLog; genericDetail(c, &m) }
func AdminPayLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	db := app.Must().DB.Model(&model.PayLog{})
	if keyword != "" {
		db = db.Where("pay_no LIKE ?", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.PayLog
	db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
func PayLogCloseHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	app.Must().DB.Model(&model.PayLog{}).Where("id = ? AND status = 0", id).Update("status", 2)
	response.OK(c, nil)
}

// ========== 地区 ==========
func RegionSaveHandler(c *gin.Context) {
	var m model.Region
	if err := c.ShouldBindJSON(&m); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if m.ID > 0 {
		app.Must().DB.Model(&m).Updates(m)
	} else {
		app.Must().DB.Create(&m)
	}
	response.OK(c, m)
}
func RegionDeleteHandler(c *gin.Context) { genericDelete(c, &model.Region{}) }

// ========== 筛选价格 ==========
func ScreeningPriceDeleteHandler(c *gin.Context) { genericDelete(c, &model.ScreeningPrice{}) }

// ========== 用户地址 ==========
func UserAddressListHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	db := app.Must().DB.Model(&model.Address{})
	if keyword != "" {
		db = db.Where("name LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.Address
	db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
func UserAddressDetailHandler(c *gin.Context) { var m model.Address; genericDetail(c, &m) }
func UserAddressSaveHandler(c *gin.Context)   { genericUpdate(c, &model.Address{}) }
func UserAddressDeleteHandler(c *gin.Context) { genericDelete(c, &model.Address{}) }

// ========== 仓库商品 ==========
func WarehouseGoodsDeleteHandler(c *gin.Context)       { genericDelete(c, &model.WarehouseGoods{}) }
func WarehouseGoodsStatusUpdateHandler(c *gin.Context) { statusUpdate(c, &model.WarehouseGoods{}) }

// ========== 邮件测试 ==========
func EmailTestHandler(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.SendEmail(req.Email, "GoShop 邮件测试", "这是一封测试邮件，如果您收到说明SMTP配置正确。"); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"msg": "测试邮件已发送"})
}

// ========== 商品浏览/收藏/购物车 管理列表 ==========
func AdminGoodsBrowseList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.BrowseHistory{}).Count(&total)
	var list []model.BrowseHistory
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
func AdminGoodsBrowseDelete(c *gin.Context) { genericDelete(c, &model.BrowseHistory{}) }

func AdminGoodsFavorList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.Favorite{}).Count(&total)
	var list []model.Favorite
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
func AdminGoodsFavorDelete(c *gin.Context) { genericDelete(c, &model.Favorite{}) }

func AdminGoodsCartList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.Cart{}).Count(&total)
	var list []model.Cart
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
func AdminGoodsCartDelete(c *gin.Context) { genericDelete(c, &model.Cart{}) }

// ========== 消息/积分/搜索记录 管理列表 ==========
func AdminMessageList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.Message{}).Count(&total)
	var list []model.Message
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminIntegralLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.PointsLog{}).Count(&total)
	var list []model.PointsLog
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminSearchHistoryList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.SearchHistory{}).Count(&total)
	var list []model.SearchHistory
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminPayRequestLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var total int64
	app.Must().DB.Model(&model.PayRequestLog{}).Count(&total)
	var list []model.PayRequestLog
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}

// ========== 促销 ==========
func PromotionUpdateHandler(c *gin.Context) { genericUpdate(c, &model.Promotion{}) }
func PromotionDeleteHandler(c *gin.Context) { genericDelete(c, &model.Promotion{}) }

// ========== 优惠券补全 ==========
func CouponUpdateHandler(c *gin.Context) { genericUpdate(c, &model.Coupon{}) }
func CouponDeleteHandler(c *gin.Context) { genericDelete(c, &model.Coupon{}) }

// ========== 物流轨迹 ==========
func LogisticsTrackHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := service.GetLogisticsTrack(uint(id))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// ========== 主题上传安装 ==========
func ThemeUploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "请选择文件")
		return
	}
	dst := fmt.Sprintf("./uploads/themes/%d_%s", time.Now().Unix(), file.Filename)
	os.MkdirAll("./uploads/themes", 0755)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	// 创建主题记录
	theme := model.ThemeData{Name: file.Filename, Data: dst}
	app.Must().DB.Create(&theme)
	response.OK(c, gin.H{"id": theme.ID, "path": dst})
}
