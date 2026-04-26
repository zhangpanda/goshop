package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 多平台登录 ==========

func PlatformLogin(c *gin.Context) {
	var req service.PlatformLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := service.PlatformLogin(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

// ========== 商品补充接口 ==========

func GoodsStockHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	spec := c.Query("spec")
	stock, _ := service.GoodsStock(uint(id), spec)
	response.OK(c, gin.H{"stock": stock})
}

func GoodsSpecDetailHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	specValues := c.Query("spec_values")
	resp, err := service.GoodsSpecDetail(uint(id), specValues)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

func GoodsScoreHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	response.OK(c, service.GoodsScore(uint(id)))
}

func HomeFloorListHandler(c *gin.Context) {
	maxCount, _ := strconv.Atoi(c.DefaultQuery("max_count", "8"))
	response.OK(c, service.HomeFloorList(maxCount))
}

func GuessYouLikeHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	response.OK(c, service.GuessYouLike(userID, limit))
}

// ========== 订单补充接口 ==========

func OrderStatusGroupTotalHandler(c *gin.Context) {
	response.OK(c, service.OrderStatusGroupTotal(c.GetUint("user_id")))
}

func OrderOperateHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := service.GetOrderDetail(c.GetUint("user_id"), uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, service.OrderOperateButtons(order))
}

func AdminOrderPayUnderLineHandler(c *gin.Context) {
	var req struct {
		OrderID uint `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.AdminOrderPayUnderLine(req.OrderID, c.GetUint("admin_id")); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

// ========== 短信/邮件日志 ==========

func SmsLogListHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, _ := service.SmsLogList(page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

func SmsLogDeleteHandler(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if len(req.IDs) == 0 {
		service.SmsLogAllDelete()
	} else {
		service.SmsLogDelete(req.IDs)
	}
	response.OK(c, nil)
}

func EmailLogListHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, _ := service.EmailLogList(page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

func EmailLogDeleteHandler(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if len(req.IDs) == 0 {
		service.EmailLogAllDelete()
	} else {
		service.EmailLogDelete(req.IDs)
	}
	response.OK(c, nil)
}
