package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func CreateOrder(c *gin.Context) {
	var req service.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	userID := c.GetUint("user_id")
	// 尝试按仓库拆单
	if orders, err := service.SplitOrderByWarehouse(userID, &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	} else if len(orders) > 0 {
		response.OK(c, orders)
		return
	}
	// 不需要拆单，走单订单创建
	order, err := service.CreateOrder(userID, &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, order)
}

func GetOrderList(c *gin.Context) {
	var req service.OrderListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := service.GetOrderList(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, resp)
}

func GetOrderDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "无效的订单ID")
		return
	}
	order, err := service.GetOrderDetail(c.GetUint("user_id"), uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, order)
}

func CancelOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "无效的订单ID")
		return
	}
	if err := service.CancelOrder(c.GetUint("user_id"), uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
