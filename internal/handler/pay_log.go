package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// CreatePayLog 创建支付日志（合并支付）
func CreatePayLog(c *gin.Context) {
	var req struct {
		OrderIDs   []uint `json:"order_ids" binding:"required,min=1"`
		PaymentID  uint   `json:"payment_id"`
		ClientType string `json:"client_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	pl, err := service.CreatePayLog(c.GetUint("user_id"), req.OrderIDs, req.PaymentID, req.ClientType)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, pl)
}

func GetPayLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, _ := service.GetPayLogList(c.GetUint("user_id"), page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminGetRefundLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, _ := service.GetRefundLogList(page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}
