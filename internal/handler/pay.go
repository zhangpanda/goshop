package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func PayOrder(c *gin.Context) {
	var req service.PayOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := service.PayOrder(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

// UnifiedPay 统一支付入口（支持多支付方式）
func UnifiedPay(c *gin.Context) {
	var req service.UnifiedPayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ClientIP = c.ClientIP()
	resp, err := service.UnifiedPay(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

// AlipayNotify 支付宝异步回调
func AlipayNotify(c *gin.Context) {
	orderNo := c.PostForm("out_trade_no")
	tradeNo := c.PostForm("trade_no")
	status := c.PostForm("trade_status")
	if status == "TRADE_SUCCESS" || status == "TRADE_FINISHED" {
		if err := service.HandlePayNotify(orderNo, tradeNo); err != nil {
			c.String(http.StatusOK, "fail")
			return
		}
	}
	c.String(http.StatusOK, "success")
}

func PayNotify(c *gin.Context) {
	handler, err := notify.NewRSANotifyHandler(global.Cfg.Wechat.MchAPIKey, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "初始化失败"})
		return
	}

	var transaction payments.Transaction
	_, err = handler.ParseNotifyRequest(c.Request.Context(), c.Request, &transaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "验签失败"})
		return
	}

	if err := service.HandlePayNotify(*transaction.OutTradeNo, *transaction.TransactionId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

func RefundOrder(c *gin.Context) {
	var req service.RefundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.RefundOrder(c.GetUint("user_id"), &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
