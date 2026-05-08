package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/zhangpanda/goshop/internal/app"
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

// AlipayNotify 支付宝异步回调（含RSA2验签）
func AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	params := make(map[string]string, len(c.Request.PostForm))
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if !service.AlipayVerifySign(params, app.Must().Cfg.Alipay.PublicKey) {
		c.String(http.StatusOK, "fail")
		return
	}
	status := params["trade_status"]
	if status == "TRADE_SUCCESS" || status == "TRADE_FINISHED" {
		if err := service.HandlePayNotify(params["out_trade_no"], params["trade_no"]); err != nil {
			c.String(http.StatusOK, "fail")
			return
		}
	}
	c.String(http.StatusOK, "success")
}

func PayNotify(c *gin.Context) {
	handler, err := notify.NewRSANotifyHandler(app.Must().Cfg.Wechat.MchAPIKey, nil)
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

	if transaction.OutTradeNo == nil || transaction.TransactionId == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "缺少交易信息"})
		return
	}

	if err := service.HandlePayNotify(*transaction.OutTradeNo, *transaction.TransactionId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// SandboxCallback 沙盒支付回调（模拟第三方回调，仅 payment.sandbox=true 时可用）
func SandboxCallback(c *gin.Context) {
	if !app.Must().Cfg.Payment.Sandbox {
		response.Fail(c, http.StatusForbidden, "沙盒模式未开启")
		return
	}
	orderNo := c.Query("order_no")
	tradeNo := c.Query("trade_no")
	if orderNo == "" {
		response.Fail(c, http.StatusBadRequest, "缺少 order_no")
		return
	}
	if tradeNo == "" {
		tradeNo = "SANDBOX_" + orderNo
	}
	if err := service.HandlePayNotify(orderNo, tradeNo); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{
		"message":  "沙盒支付成功",
		"order_no": orderNo,
		"trade_no": tradeNo,
	})
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
