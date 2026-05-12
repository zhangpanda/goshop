package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务错误码定义。前端可按 code 做差异化处理。
// 约定：0=成功，负数=通用错误，正数=业务错误（按模块分段）。
const (
	CodeSuccess = 0
	CodeFail    = -1 // 通用失败（兼容旧接口）

	// 1xxx: 认证/权限
	CodeUnauthorized = 1001
	CodeForbidden    = 1002
	CodeTokenExpired = 1003

	// 2xxx: 参数校验
	CodeBadParam    = 2001
	CodeMissingID   = 2002
	CodeInvalidJSON = 2003

	// 3xxx: 商品
	CodeGoodsNotFound = 3001
	CodeGoodsOffline  = 3002
	CodeStockInsuf    = 3003

	// 4xxx: 订单
	CodeOrderNotFound    = 4001
	CodeOrderStatusErr   = 4002
	CodeCartEmpty        = 4003
	CodeCouponInvalid    = 4004
	CodeRefundDuplicate  = 4005
	CodeRefundStatusErr  = 4006

	// 5xxx: 支付
	CodePayNotConfigured = 5001
	CodePayFailed        = 5002
	CodeWalletInsuf      = 5003

	// 6xxx: 用户
	CodeUserNotFound  = 6001
	CodeUserDisabled  = 6002
	CodePasswordWrong = 6003
	CodeVerifyCode    = 6004

	// 9xxx: 系统
	CodeRateLimit = 9001
	CodeInternal  = 9999
)

// BizError 业务错误，携带错误码。
type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

// NewBizError 创建业务错误。
func NewBizError(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

// FailCode 返回带业务错误码的失败响应。HTTP 状态码根据 code 范围自动推断。
func FailCode(c *gin.Context, code int, msg string) {
	httpCode := http.StatusBadRequest
	switch {
	case code >= 1001 && code <= 1003:
		httpCode = http.StatusUnauthorized
	case code == CodeForbidden:
		httpCode = http.StatusForbidden
	case code == CodeRateLimit:
		httpCode = http.StatusTooManyRequests
	case code == CodeInternal:
		httpCode = http.StatusInternalServerError
	}
	c.JSON(httpCode, Response{Code: code, Msg: msg})
}

// FailBiz 从 BizError 返回响应。
func FailBiz(c *gin.Context, err *BizError) {
	FailCode(c, err.Code, err.Msg)
}
