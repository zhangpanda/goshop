package shopxo

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/handler"
	"github.com/zhangpanda/goshop/pkg/auth"
	"github.com/zhangpanda/goshop/pkg/response"
)

// SetupShopXOCompat 注册 /api.php 兼容路由（对照 shopxo-uniapp 常见 s=controller/action 形态）。
// 请求示例: /api.php?s=controller/action&token=xxx&ajax=ajax
// ShopXO/PHP 支付方式字段（config.payment 类名）仅在同包 payment_key.go 解析，避免污染 internal/service。
func SetupShopXOCompat(r *gin.Engine) {
	r.Any("/api.php", shopxoDispatch)
}

func shopxoDispatch(c *gin.Context) {
	s := c.Query("s")
	s = strings.TrimPrefix(s, "/")
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		response.Fail(c, http.StatusBadRequest, "invalid route")
		return
	}
	ctrl := strings.ToLower(parts[0])
	action := strings.ToLower(parts[1])

	// token鉴权：从query参数取token，设置到context
	token := c.Query("token")
	if token != "" {
		userID := parseTokenToUserID(token)
		if userID > 0 {
			c.Set("user_id", userID)
		}
	}

	key := ctrl + "/" + action
	h, ok := routeMap[key]
	if !ok {
		response.Fail(c, http.StatusNotFound, "接口不存在: "+s)
		return
	}

	// 需要登录的路由，拒绝未认证请求
	if authRequiredRoutes[key] && c.GetUint("user_id") == 0 {
		response.Fail(c, http.StatusUnauthorized, "请先登录")
		return
	}

	h(c)
}

// authRequiredRoutes 需要登录才能访问的路由
var authRequiredRoutes = map[string]bool{
	"user/tokenuserinfo": true, "user/center": true,
	"user/appminiuserinfo": true, "user/appmobilebind": true, "user/appemailbind": true,
	"user/onekeyusermobilebind": true,
	"cart/index":                true, "cart/save": true, "cart/delete": true,
	"buy/index": true, "buy/add": true,
	"order/index": true, "order/detail": true, "order/pay": true, "order/paycheck": true,
	"order/cancel": true, "order/collect": true, "order/delete": true, "order/comments": true,
	"orderaftersale/index": true, "orderaftersale/detail": true, "orderaftersale/create": true,
	"orderaftersale/delivery": true, "orderaftersale/cancel": true,
	"useraddress/index": true, "useraddress/detail": true, "useraddress/save": true,
	"useraddress/delete": true, "useraddress/setdefault": true, "useraddress/outsystemadd": true,
	"personal/index": true, "personal/save": true, "personal/useravatarupload": true,
	"safety/loginpwdupdate": true, "safety/logout": true,
	"goods/favor":          true,
	"usergoodsfavor/index": true, "usergoodsfavor/cancel": true,
	"usergoodsbrowse/index": true, "usergoodsbrowse/delete": true,
	"usergoodscomments/save": true, "usergoodscomments/delete": true,
	"userintegral/index": true, "message/index": true,
	"forminputdata/save": true, "forminputdata/delete": true, "forminputdata/detail": true,
	"paylog/index": true, "paylog/detail": true,
	"order/commentssave": true,
}

func parseTokenToUserID(token string) uint {
	claims, err := parseJWTToken(token)
	if err != nil || claims == nil {
		return 0
	}
	return claims.UserID
}

func parseJWTToken(token string) (*auth.Claims, error) {
	return auth.ParseToken(token, app.Must().Cfg.JWT.Secret)
}

// routeMap 兼容层 controller/action -> 本仓库 handler（命名沿历史习惯，非主张与 ShopXO 逐字节一致）
var routeMap = map[string]gin.HandlerFunc{
	// ===== base =====
	"base/common": sxCommon,

	// ===== index =====
	"index/index": sxIndexIndex,

	// ===== user =====
	"user/login":                   sxUserLogin,
	"user/reg":                     sxUserReg,
	"user/forgetpwd":               sxUserForgetPwd,
	"user/tokenuserinfo":           sxTokenUserInfo,
	"user/center":                  sxUserCenter,
	"user/appminiuserauth":         sxAppMiniUserAuth,
	"user/appminiuserinfo":         sxAppMiniUserInfo,
	"user/appmobilebind":           sxAppMobileBind,
	"user/appemailbind":            sxAppEmailBind,
	"user/onekeyusermobilebind":    sxOnekeyMobileBind,
	"user/onekeyusermobiledecrypt": sxOnekeyMobileDecrypt,
	"user/userbasereg":             sxUserBaseReg,
	"user/loginverifysend":         sxLoginVerifySend,
	"user/regverifysend":           sxRegVerifySend,
	"user/forgetpwdverifysend":     sxForgetPwdVerifySend,
	"user/appmobilebindverifysend": sxMobileBindVerifySend,
	"user/appemailbindverifysend":  sxEmailBindVerifySend,
	"user/userverifyentry":         sxUserVerifyEntry,

	// ===== article（对照 shopxo-uniapp 文章类页面）=====
	"article/index":    sxArticleIndex,
	"article/datalist": sxArticleDataList,
	"article/detail":   sxArticleDetail,

	// ===== goods =====
	"goods/detail":     sxGoodsDetail,
	"goods/category":   sxGoodsCategory,
	"goods/favor":      sxGoodsFavor,
	"goods/spectype":   sxGoodsSpecType,
	"goods/specdetail": sxGoodsSpecDetail,
	"goods/stock":      sxGoodsStock,
	"goods/goodsscore": sxGoodsScore,
	"goods/comments":   sxGoodsComments,

	// ===== search =====
	"search/index":    sxSearchIndex,
	"search/datalist": sxSearchDataList,
	"search/start":    sxSearchStart,

	// ===== cart =====
	"cart/index":  sxCartIndex,
	"cart/save":   sxCartSave,
	"cart/delete": sxCartDelete,
	"cart/stock":  sxCartStock,

	// ===== buy =====
	"buy/index": sxBuyIndex,
	"buy/add":   sxBuyAdd,

	// ===== order =====
	"order/index":        sxOrderIndex,
	"order/detail":       sxOrderDetail,
	"order/pay":          sxOrderPay,
	"order/paycheck":     sxOrderPayCheck,
	"order/cancel":       sxOrderCancel,
	"order/collect":      sxOrderCollect,
	"order/delete":       sxOrderDelete,
	"order/comments":     sxOrderComments,
	"order/commentssave": sxOrderCommentSave,

	// ===== orderaftersale =====
	"orderaftersale/index":    sxAftersaleIndex,
	"orderaftersale/detail":   sxAftersaleDetail,
	"orderaftersale/create":   handler.AftersaleCreate,
	"orderaftersale/delivery": sxAftersaleDelivery,
	"orderaftersale/cancel":   sxAftersaleCancel,

	// ===== useraddress =====
	"useraddress/index":        sxAddressIndex,
	"useraddress/detail":       sxAddressDetail,
	"useraddress/save":         sxAddressSave,
	"useraddress/delete":       sxAddressDelete,
	"useraddress/setdefault":   sxAddressSetDefault,
	"useraddress/extraction":   sxAddressExtraction,
	"useraddress/outsystemadd": sxAddressOutSystemAdd,

	// ===== personal =====
	"personal/index":            sxPersonalIndex,
	"personal/save":             sxPersonalSave,
	"personal/useravatarupload": handler.Upload,

	// ===== safety =====
	"safety/loginpwdupdate": sxLoginPwdUpdate,
	"safety/logout":         sxLogout,

	// ===== usergoodsfavor =====
	"usergoodsfavor/index":  sxFavorIndex,
	"usergoodsfavor/cancel": sxFavorCancel,

	// ===== usergoodsbrowse =====
	"usergoodsbrowse/index":  sxBrowseIndex,
	"usergoodsbrowse/delete": sxBrowseDelete,

	// ===== usergoodscomments =====
	"usergoodscomments/index":  sxCommentsIndex,
	"usergoodscomments/detail": sxCommentsDetail,
	"usergoodscomments/save":   sxCommentsSave,
	"usergoodscomments/delete": sxCommentsDelete,

	// ===== userintegral =====
	"userintegral/index": sxIntegralIndex,

	// ===== message =====
	"message/index": sxMessageIndex,

	// ===== region =====
	"region/index":    sxRegionIndex,
	"region/all":      sxRegionAll,
	"region/codedata": sxRegionCodeData,

	// ===== agreement =====
	"agreement/index": sxAgreementIndex,

	// ===== diy =====
	"diy/index": sxDiyIndex,

	// ===== design / customview（可视化装修、自定义页）=====
	"design/index":     sxDesignIndex,
	"customview/index": sxCustomviewIndex,

	// ===== forminput =====
	"forminput/index":       sxFormInputIndex,
	"forminput/verifysend":  sxFormInputVerifySend,
	"forminput/verifyentry": sxFormInputVerifyEntry,

	// ===== forminputdata =====
	"forminputdata/index":  sxFormInputDataIndex,
	"forminputdata/save":   sxFormInputDataSave,
	"forminputdata/delete": sxFormInputDataDelete,
	"forminputdata/detail": sxFormInputDataDetail,

	// ===== paylog（支付日志 / 订单明细式列表）=====
	"paylog/index":  sxPaylogIndex,
	"paylog/detail": sxPaylogDetail,

	// ===== ueditor（富文本图片上传，shopxo-uniapp 常用 action=uploadimage）=====
	"ueditor/index": sxUeditorIndex,

	// ===== cashier（微信小程序收银台，对照 shopxo-uniapp cashier/paydata 约定）=====
	"cashier/paydata": sxCashierPayData,
}
