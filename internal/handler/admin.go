package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 商品管理 ==========

func AdminUpdateGoods(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		service.GoodsReq
		SKUs []struct {
			Name   string `json:"name"`
			Price  int64  `json:"price"`
			Stock  int    `json:"stock"`
			Image  string `json:"image"`
			Specs  string `json:"specs"`
			Coding string `json:"coding"`
		} `json:"skus"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	tx := global.DB.Begin()
	if err := tx.Model(&model.Goods{}).Where("id = ?", id).Updates(map[string]interface{}{
		"category_id": req.CategoryID, "title": req.Title, "subtitle": req.Subtitle,
		"main_image": req.MainImage, "images": req.Images, "detail": req.Detail,
	}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	// 更新SKU：先删后建
	if len(req.SKUs) > 0 {
		if err := tx.Where("goods_id = ?", id).Delete(&model.GoodsSKU{}).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		for _, s := range req.SKUs {
			if err := tx.Create(&model.GoodsSKU{
				GoodsID: uint(id), Name: s.Name, Price: s.Price, Stock: s.Stock,
				Image: s.Image, Specs: s.Specs, Coding: s.Coding, Status: 1,
			}).Error; err != nil {
				tx.Rollback()
				response.Fail(c, http.StatusInternalServerError, err.Error())
				return
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		// Tx is invalid after Commit returns; do not Rollback.
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

func AdminDeleteGoods(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.GoodsDeleteFull(uint(id)); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

func AdminToggleGoodsStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status int8 `json:"status"`
	}
	c.ShouldBindJSON(&req)
	global.DB.Model(&model.Goods{}).Where("id = ?", id).Update("status", req.Status)
	response.OK(c, nil)
}

// ========== 订单管理 ==========

func AdminGetOrders(c *gin.Context) {
	var req service.OrderListReq
	c.ShouldBindQuery(&req)
	db := global.DB.Model(&model.Order{})
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		db = db.Where("order_no LIKE ?", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.Order
	db.Preload("Items").Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminUpdateOrderRemark(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Remark string `json:"remark"`
	}
	c.ShouldBindJSON(&req)
	global.DB.Model(&model.Order{}).Where("id = ?", id).Update("remark", req.Remark)
	response.OK(c, nil)
}

// ========== 用户管理 ==========

func AdminGetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	idsStr := c.Query("ids")

	db := global.DB.Model(&model.User{})
	if idsStr != "" {
		var ids []uint
		for _, s := range strings.Split(idsStr, ",") {
			if id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); err == nil && id > 0 {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) > 0 {
			db = db.Where("id IN ?", ids)
		}
	}
	if keyword != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR phone LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	db.Count(&total)
	var list []model.User
	db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminUpdateUserStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Status int8 `json:"status"`
	}
	c.ShouldBindJSON(&req)
	global.DB.Model(&model.User{}).Where("id = ?", id).Update("status", req.Status)
	response.OK(c, nil)
}

// AdminDeleteUserHandler 禁用用户（非物理删除，与前台订单数据兼容）。
func AdminDeleteUserHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.AdminDisableUser(uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

// ========== 分类管理 ==========

func AdminUpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	global.DB.Model(&model.Category{}).Where("id = ?", id).Updates(map[string]interface{}{
		"parent_id": req.ParentID, "name": req.Name, "icon": req.Icon, "sort": req.Sort,
	})
	service.InvalidateCategoryCache()
	response.OK(c, nil)
}

func AdminDeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	// 检查是否有子分类
	var count int64
	global.DB.Model(&model.Category{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, "请先删除子分类")
		return
	}
	global.DB.Delete(&model.Category{}, id)
	service.InvalidateCategoryCache()
	response.OK(c, nil)
}

// ========== 数据统计 ==========

func AdminDashboard(c *gin.Context) {
	var userCount, goodsCount, orderCount int64
	var todayOrderCount int64
	var todaySales int64

	global.DB.Model(&model.User{}).Count(&userCount)
	global.DB.Model(&model.Goods{}).Where("status = 1").Count(&goodsCount)
	global.DB.Model(&model.Order{}).Count(&orderCount)

	today := time.Now().Format("2006-01-02")
	global.DB.Model(&model.Order{}).Where("DATE(created_at) = ? AND status > 0", today).Count(&todayOrderCount)
	global.DB.Model(&model.Order{}).Where("DATE(created_at) = ? AND status > 0", today).
		Select("COALESCE(SUM(pay_amount),0)").Scan(&todaySales)

	// 近7天销售趋势
	type DaySales struct {
		Date  string `json:"date"`
		Sales int64  `json:"sales"`
		Count int64  `json:"count"`
	}
	var trend []DaySales
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var ds DaySales
		ds.Date = d
		global.DB.Model(&model.Order{}).Where("DATE(created_at) = ? AND status > 0", d).
			Select("COALESCE(SUM(pay_amount),0) as sales, COUNT(*) as count").Scan(&ds)
		ds.Date = d
		trend = append(trend, ds)
	}

	response.OK(c, gin.H{
		"user_count":        userCount,
		"goods_count":       goodsCount,
		"order_count":       orderCount,
		"today_order_count": todayOrderCount,
		"today_sales":       todaySales,
		"trend":             trend,
	})
}
