package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/auth"
	"github.com/zhangpanda/goshop/pkg/response"
)

// SetupShopXOCompat 注册ShopXO兼容路由
// uni-app请求格式: /api.php?s=controller/action&token=xxx&ajax=ajax
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
	"forminputdata/save": true, "forminputdata/delete": true,
}

func parseTokenToUserID(token string) uint {
	claims, err := parseJWTToken(token)
	if err != nil || claims == nil {
		return 0
	}
	return claims.UserID
}

func parseJWTToken(token string) (*auth.Claims, error) {
	return auth.ParseToken(token, global.Cfg.JWT.Secret)
}

// routeMap ShopXO controller/action -> Go handler
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
	"order/index":    sxOrderIndex,
	"order/detail":   sxOrderDetail,
	"order/pay":      sxOrderPay,
	"order/paycheck": sxOrderPayCheck,
	"order/cancel":   sxOrderCancel,
	"order/collect":  sxOrderCollect,
	"order/delete":   sxOrderDelete,
	"order/comments": sxOrderComments,

	// ===== orderaftersale =====
	"orderaftersale/index":    sxAftersaleIndex,
	"orderaftersale/detail":   sxAftersaleDetail,
	"orderaftersale/create":   sxAftersaleCreate,
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
	"personal/useravatarupload": sxUserAvatarUpload,

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

	// ===== forminput =====
	"forminput/index":      sxFormInputIndex,
	"forminput/verifysend": sxFormInputVerifySend,

	// ===== forminputdata =====
	"forminputdata/index":  sxFormInputDataIndex,
	"forminputdata/save":   sxFormInputDataSave,
	"forminputdata/delete": sxFormInputDataDelete,

	// ===== cashier（微信小程序收银台，与 ShopXO Cashier::PayData 一致）=====
	"cashier/paydata": sxCashierPayData,
}

// ===== handler实现 =====

func sxCommon(c *gin.Context) {
	// 构造ShopXO base/common 完整返回结构
	config := map[string]interface{}{
		"common_site_type":                           service.GetConfig("common_site_type"),
		"common_shop_notice":                         service.GetConfig("common_shop_notice"),
		"common_app_is_enable_search":                1,
		"common_app_is_online_service":               0,
		"common_app_customer_service_tel":            service.GetConfig("app_customer_service_tel"),
		"common_app_h5_url":                          service.GetConfig("common_app_h5_url"),
		"common_order_close_limit_time":              30,
		"common_order_is_booking":                    0,
		"common_verify_expire_time":                  600,
		"common_verify_interval_time":                60,
		"common_img_verify_state":                    0,
		"home_user_login_type":                       []string{"username"},
		"home_user_reg_type":                         []string{"username"},
		"home_user_login_img_verify_state":           0,
		"home_user_register_img_verify_state":        0,
		"common_user_is_mandatory_bind_mobile":       0,
		"common_user_verify_bind_mobile_list":        []string{},
		"common_user_onekey_bind_mobile_list":        []string{},
		"common_user_address_platform_import_list":   []string{},
		"common_app_is_weixin_force_user_base":       0,
		"common_app_user_base_popup_pages":           []string{},
		"common_app_user_base_popup_client":          []string{},
		"home_site_name":                             service.GetConfig("home_site_name"),
		"home_site_logo":                             service.GetConfig("home_site_logo"),
		"home_site_logo_app":                         service.GetConfig("home_site_logo_app"),
		"home_site_logo_square":                      service.GetConfig("home_site_logo_square"),
		"home_search_is_brand":                       1,
		"home_search_is_category":                    1,
		"home_search_is_price":                       1,
		"home_search_is_params":                      1,
		"home_search_is_spec":                        1,
		"home_search_limit_number":                   20,
		"home_user_address_map_status":               0,
		"home_user_address_idcard_status":            0,
		"home_is_enable_order_bulk_pay":              0,
		"home_use_multilingual_status":               0,
		"common_is_goods_detail_show_comments":       1,
		"common_is_goods_detail_show_seeing_you":     1,
		"common_is_goods_detail_show_guess_you_like": 1,
		"common_map_type":                            "baidu",
		"category_show_level":                        0,
		"common_goods_cover_size_type":               0,
		"home_is_enable_userregister_agreement":      0,
		"agreement_userregister_url":                 "/api.php?s=agreement/index&name=userregister",
		"agreement_userprivacy_url":                  "/api.php?s=agreement/index&name=userprivacy",
		"agreement_userlogout_url":                   "/api.php?s=agreement/index&name=userlogout",
	}
	// 底部导航
	tabbar, _ := service.AppTabbarList()
	// 快捷导航
	quickNav, _ := service.QuickNavList()
	// 货币符号
	currCfg := service.GetCurrencyConfig()

	data := map[string]interface{}{
		"status":          1,
		"config":          config,
		"app_tabbar":      tabbar,
		"currency_symbol": currCfg.Symbol,
		"quick_nav":       quickNav,
		"plugins_base":    []interface{}{},
		"site_info_data":  map[string]interface{}{},
	}
	response.OK(c, data)
}

func sxIndexIndex(c *gin.Context) {
	// 尝试DIY模式
	diyData := service.AppClientHomeDiyData()
	if diyData != nil {
		uid := c.GetUint("user_id")
		cartN := service.GoodsCartTotal(uid)
		response.OK(c, map[string]interface{}{
			"data_mode": 3, "data_list": diyData, "cart_total": map[string]int64{"buy_number": cartN},
		})
		return
	}
	// 非DIY模式：返回楼层+轮播+导航
	slides, _ := service.SlideList()
	homeNav, _ := service.AppHomeNavList()
	floors := service.HomeFloorList(8)
	userID := c.GetUint("user_id")
	cartTotal := service.GoodsCartTotal(userID)
	response.OK(c, map[string]interface{}{
		"data_mode":     0,
		"navigation":    homeNav,
		"banner_list":   slides,
		"data_list":     floors,
		"cart_total":    map[string]int64{"buy_number": cartTotal},
		"message_total": service.UnreadCount(userID),
	})
}

func sxUserLogin(c *gin.Context) {
	var body struct {
		Accounts string `json:"accounts" form:"accounts"`
		Pwd      string `json:"pwd" form:"pwd"`
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	c.ShouldBind(&body)
	req := service.LoginReq{
		Username: body.Username,
		Password: body.Password,
	}
	if body.Accounts != "" {
		req.Username = body.Accounts
	}
	if body.Pwd != "" {
		req.Password = body.Pwd
	}
	if req.Username == "" || req.Password == "" {
		response.Fail(c, http.StatusBadRequest, "请输入账号和密码")
		return
	}
	resp, err := service.Login(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

func sxUserReg(c *gin.Context) {
	var body struct {
		Accounts string `json:"accounts" form:"accounts"`
		Pwd      string `json:"pwd" form:"pwd"`
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	c.ShouldBind(&body)
	req := service.RegisterReq{
		Username: body.Username,
		Password: body.Password,
	}
	if body.Accounts != "" {
		req.Username = body.Accounts
	}
	if body.Pwd != "" {
		req.Password = body.Pwd
	}
	if req.Username == "" || req.Password == "" {
		response.Fail(c, http.StatusBadRequest, "请输入账号和密码")
		return
	}
	user, err := service.Register(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, user)
}

func sxUserForgetPwd(c *gin.Context) {
	var req service.ForgetPwdReq
	c.ShouldBind(&req)
	if err := service.ForgetPassword(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func sxTokenUserInfo(c *gin.Context) {
	response.OK(c, service.LoginUserInfo(c.GetUint("user_id")))
}

func sxUserCenter(c *gin.Context) {
	response.OK(c, service.UserInfo(c.GetUint("user_id")))
}

func sxAppMiniUserAuth(c *gin.Context) {
	var req service.WxLoginReq
	c.ShouldBind(&req)
	resp, err := service.WxLogin(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

func sxAppMiniUserInfo(c *gin.Context) {
	response.OK(c, service.AppUserInfoHandle(service.LoginUserInfo(c.GetUint("user_id"))))
}

func sxAppMobileBind(c *gin.Context) {
	var req service.BindMobileReq
	c.ShouldBind(&req)
	if err := service.BindMobile(c.GetUint("user_id"), &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func sxAppEmailBind(c *gin.Context) {
	email := c.PostForm("email")
	code := c.PostForm("code")
	if err := service.AppEmailBind(c.GetUint("user_id"), email, code); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func sxOnekeyMobileBind(c *gin.Context) {
	mobile := c.PostForm("mobile")
	service.AppMiniOnekeyUserMobileBind(c.GetUint("user_id"), mobile)
	response.OK(c, nil)
}

func sxOnekeyMobileDecrypt(c *gin.Context) { response.OK(c, nil) }
func sxUserBaseReg(c *gin.Context)         { sxUserReg(c) }
func sxLoginVerifySend(c *gin.Context) {
	service.LoginVerifySend(c.PostForm("accounts"))
	response.OK(c, nil)
}
func sxRegVerifySend(c *gin.Context) {
	service.RegVerifySend(c.PostForm("accounts"))
	response.OK(c, nil)
}
func sxForgetPwdVerifySend(c *gin.Context) {
	service.ForgetPwdVerifySend(c.PostForm("accounts"))
	response.OK(c, nil)
}
func sxMobileBindVerifySend(c *gin.Context) {
	service.AppMobileBindVerifySend(c.PostForm("mobile"))
	response.OK(c, nil)
}
func sxEmailBindVerifySend(c *gin.Context) {
	service.AppEmailBindVerifySend(c.PostForm("email"))
	response.OK(c, nil)
}

func sxGoodsDetail(c *gin.Context) {
	id := getID(c)
	goods, err := service.GetGoodsDetail(id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "商品不存在")
		return
	}
	response.OK(c, goods)
}

func sxGoodsCategory(c *gin.Context) {
	cats := service.GoodsCategoryAll()
	response.OK(c, cats)
}

func sxGoodsFavor(c *gin.Context) {
	id := getID(c)
	added, _ := service.ToggleFavorite(c.GetUint("user_id"), id)
	response.OK(c, map[string]interface{}{"is_favorite": added})
}

func sxGoodsSpecType(c *gin.Context) {
	id := getID(c)
	response.OK(c, service.GoodsSpecType(id))
}

func sxGoodsSpecDetail(c *gin.Context) {
	id := getID(c)
	spec := c.Query("spec")
	resp, err := service.GoodsSpecDetail(id, spec)
	if err != nil {
		response.OK(c, map[string]interface{}{"spec": nil, "sku": nil})
		return
	}
	response.OK(c, resp)
}

func sxGoodsStock(c *gin.Context) {
	id := getID(c)
	stock, _ := service.GoodsStock(id, c.Query("spec"))
	response.OK(c, map[string]int{"stock": stock})
}

func sxGoodsScore(c *gin.Context) {
	response.OK(c, service.GoodsScore(getID(c)))
}

func sxGoodsComments(c *gin.Context) {
	id := getID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetGoodsReviews(id, page, 10)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}

func sxSearchIndex(c *gin.Context) {
	var req service.GoodsListReq
	c.ShouldBindQuery(&req)
	if wd := c.Query("wd"); wd != "" {
		req.Keyword = wd
	}
	if kw := c.Query("keywords"); kw != "" {
		req.Keyword = kw
	}
	resp, _ := service.GetGoodsList(&req)
	response.OK(c, resp)
}
func sxSearchDataList(c *gin.Context) {
	var req service.GoodsListReq
	c.ShouldBindQuery(&req)
	if wd := c.Query("wd"); wd != "" {
		req.Keyword = wd
	}
	resp, _ := service.GetGoodsList(&req)
	response.OK(c, resp)
}
func sxSearchStart(c *gin.Context) { response.OK(c, service.SearchStartData()) }

func sxCartIndex(c *gin.Context) {
	list, _ := service.GetCartList(c.GetUint("user_id"))
	response.OK(c, list)
}
func sxCartSave(c *gin.Context) {
	var req service.AddCartReq
	c.ShouldBind(&req)
	cart, err := service.AddCart(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, cart)
}
func sxCartDelete(c *gin.Context) {
	ids := parseUintSlice(c.PostForm("ids"))
	service.DeleteCart(c.GetUint("user_id"), ids)
	response.OK(c, nil)
}
func sxCartStock(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	stock, _ := service.GoodsCartStock(uint(id))
	response.OK(c, map[string]int{"stock": stock})
}

func sxBuyIndex(c *gin.Context) {
	ids := parseUintSlice(c.Query("ids"))
	buyType := c.DefaultQuery("buy_type", "cart")
	resp, err := service.BuyOrderInit(c.GetUint("user_id"), ids, buyType)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}
func sxBuyAdd(c *gin.Context) {
	// ShopXO buy/add: 直接用 goods_id+sku_id 下单，需要先加购物车
	var req struct {
		GoodsID   uint `json:"goods_id" form:"goods_id"`
		SKUID     uint `json:"sku_id" form:"sku_id"`
		Stock     int  `json:"stock" form:"stock"`
		AddressID uint `json:"address_id" form:"address_id"`
	}
	c.ShouldBind(&req)
	if req.Stock == 0 {
		req.Stock = 1
	}
	userID := c.GetUint("user_id")
	cart, err := service.AddCart(userID, &service.AddCartReq{GoodsID: req.GoodsID, SKUID: req.SKUID, Quantity: req.Stock})
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	addrID := req.AddressID
	order, err := service.CreateOrder(userID, &service.CreateOrderReq{AddressID: &addrID, CartIDs: []uint{cart.ID}})
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, order)
}

func sxOrderIndex(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req service.OrderListReq
	if err := c.ShouldBind(&req); err != nil {
		req.Page = 1
		req.PageSize = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	payload, err := service.ShopXOOrderIndexPayload(userID, &req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, payload)
}

func sxOrderDetail(c *gin.Context) {
	id := getID(c)
	order, err := service.GetOrderDetail(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	payRows, _ := service.ShopXOUserPaymentRows()
	if payRows == nil {
		payRows = []map[string]interface{}{}
	}
	detail := service.ShopXOOrderDetailView(order)
	response.OK(c, map[string]interface{}{
		"data":               detail,
		"operate":            service.OrderOperateButtons(order),
		"steps":              service.OrderStepData(order),
		"payment_list":       payRows,
		"default_payment_id": service.DefaultPaymentIDForShopXO(),
		"status_tips":        "",
		"site_fictitious":    nil,
	})
}

func sxOrderPay(c *gin.Context) {
	userID := c.GetUint("user_id")
	var body struct {
		IDs       string `form:"ids" json:"ids"`
		ID        string `form:"id" json:"id"`
		PaymentID uint   `form:"payment_id" json:"payment_id"`
		OpenID    string `form:"openid" json:"openid"`
		ReturnURL string `form:"return_url" json:"return_url"`
	}
	_ = c.ShouldBind(&body)
	idsStr := strings.TrimSpace(body.IDs)
	if idsStr == "" {
		idsStr = strings.TrimSpace(body.ID)
	}
	ids := parseUintSlice(idsStr)
	if len(ids) == 0 {
		response.Fail(c, http.StatusBadRequest, "请选择订单")
		return
	}
	if body.PaymentID == 0 {
		response.Fail(c, http.StatusBadRequest, "请选择支付方式")
		return
	}
	payload, outerMsg, err := service.ShopXOCompatUnifiedPay(userID, ids, body.PaymentID, body.OpenID, body.ReturnURL, c.ClientIP())
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if outerMsg != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": *outerMsg, "data": payload})
		return
	}
	response.OK(c, payload)
}
func sxOrderPayCheck(c *gin.Context) {
	id := getID(c)
	_, err := service.OrderPayCheck(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

/**
 * sxCashierPayData ShopXO cashier/paydata：小程序内用 wx.login 的 code 换 openid 后拉取 JSAPI 支付参数（需先通过 order/pay 创建 PayLog）。
 */
func sxCashierPayData(c *gin.Context) {
	var body struct {
		AuthCode string `form:"authcode" json:"authcode"`
		OrderNo  string `form:"order_no" json:"order_no"`
	}
	_ = c.ShouldBind(&body)
	if body.AuthCode == "" {
		body.AuthCode = c.Query("authcode")
	}
	if body.OrderNo == "" {
		body.OrderNo = c.Query("order_no")
	}
	payload, err := service.ShopXOCashierPayData(body.AuthCode, body.OrderNo)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, payload)
}

func sxOrderCancel(c *gin.Context) {
	id := getID(c)
	if err := service.CancelOrder(c.GetUint("user_id"), id); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
func sxOrderCollect(c *gin.Context) {
	id := getID(c)
	if err := service.ConfirmReceive(c.GetUint("user_id"), id); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
func sxOrderDelete(c *gin.Context) {
	id := getID(c)
	service.DeleteOrder(c.GetUint("user_id"), id)
	response.OK(c, nil)
}
func sxOrderComments(c *gin.Context) {
	id := getID(c)
	items := service.OrderItemList(id)
	response.OK(c, items)
}

func sxAftersaleIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetAftersaleList(c.GetUint("user_id"), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxAftersaleDetail(c *gin.Context) {
	id := getID(c)
	as, err := service.GetAftersaleDetail(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, service.OrderAftersaleDetailData(as.ID))
}
func sxAftersaleCreate(c *gin.Context) { AftersaleCreate(c) }
func sxAftersaleDelivery(c *gin.Context) {
	id := getID(c)
	var req service.AftersaleDeliveryReq
	c.ShouldBind(&req)
	if err := service.AftersaleDelivery(c.GetUint("user_id"), id, &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
func sxAftersaleCancel(c *gin.Context) {
	id := getID(c)
	service.AftersaleCancel(c.GetUint("user_id"), id)
	response.OK(c, nil)
}

func sxAddressIndex(c *gin.Context) {
	list, _ := service.GetAddressList(c.GetUint("user_id"))
	response.OK(c, list)
}
func sxAddressDetail(c *gin.Context) {
	id := getID(c)
	addr, err := service.GetAddress(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, addr)
}
func sxAddressSave(c *gin.Context) {
	var req service.AddressReq
	c.ShouldBind(&req)
	id, _ := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if id > 0 {
		service.UpdateAddress(c.GetUint("user_id"), uint(id), &req)
	} else {
		service.CreateAddress(c.GetUint("user_id"), &req)
	}
	response.OK(c, nil)
}
func sxAddressDelete(c *gin.Context) {
	id := getID(c)
	service.DeleteAddress(c.GetUint("user_id"), id)
	response.OK(c, nil)
}
func sxAddressSetDefault(c *gin.Context) {
	// 设为默认
	id := getID(c)
	service.UpdateAddress(c.GetUint("user_id"), id, &service.AddressReq{IsDefault: true})
	response.OK(c, nil)
}
func sxAddressExtraction(c *gin.Context) {
	response.OK(c, service.GetSelfExtractionAddressList())
}
func sxAddressOutSystemAdd(c *gin.Context) {
	var req service.AddressReq
	c.ShouldBind(&req)
	service.CreateAddress(c.GetUint("user_id"), &req)
	response.OK(c, nil)
}

func sxPersonalIndex(c *gin.Context) {
	response.OK(c, service.LoginUserInfo(c.GetUint("user_id")))
}
func sxPersonalSave(c *gin.Context) {
	var req service.PersonalSaveReq
	c.ShouldBind(&req)
	service.PersonalSave(c.GetUint("user_id"), &req)
	response.OK(c, nil)
}
func sxUserAvatarUpload(c *gin.Context) { Upload(c) }

func sxLoginPwdUpdate(c *gin.Context) { UpdatePassword(c) }
func sxLogout(c *gin.Context)         { response.OK(c, nil) }

func sxFavorIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetFavorites(c.GetUint("user_id"), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxFavorCancel(c *gin.Context) {
	id := getID(c)
	service.ToggleFavorite(c.GetUint("user_id"), id)
	response.OK(c, nil)
}

func sxBrowseIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetBrowseHistory(c.GetUint("user_id"), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxBrowseDelete(c *gin.Context) {
	ids := parseUintSlice(c.PostForm("ids"))
	service.GoodsBrowseDelete(ids, c.GetUint("user_id"))
	response.OK(c, nil)
}

func sxCommentsIndex(c *gin.Context)  { sxGoodsComments(c) }
func sxCommentsDetail(c *gin.Context) { response.OK(c, nil) }
func sxCommentsSave(c *gin.Context)   { CreateReview(c) }
func sxCommentsDelete(c *gin.Context) {
	id := getID(c)
	if err := service.GoodsCommentsDeleteForUser(id, c.GetUint("user_id")); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func sxIntegralIndex(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetPointsLog(c.GetUint("user_id"), page, 20)
	integral := service.UserIntegral(c.GetUint("user_id"))
	response.OK(c, map[string]interface{}{"total": total, "data": list, "integral": integral})
}

func sxMessageIndex(c *gin.Context) { GetMessages(c) }

func sxRegionIndex(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.DefaultQuery("parent_id", "0"), 10, 64)
	list, _ := service.GetRegionList(uint(pid))
	response.OK(c, list)
}
func sxRegionAll(c *gin.Context) { response.OK(c, service.RegionAll()) }
func sxRegionCodeData(c *gin.Context) {
	id := getID(c)
	response.OK(c, service.RegionCodeData(id))
}

func sxAgreementIndex(c *gin.Context) {
	name := c.Query("name")
	response.OK(c, service.AgreementGet(name))
}

func sxDiyIndex(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id > 0 {
		data, _ := service.DiyApiCustomInit(uint(id))
		response.OK(c, data)
	} else {
		response.OK(c, service.AppClientHomeDiyData())
	}
}

func sxFormInputIndex(c *gin.Context) {
	id := getID(c)
	response.OK(c, service.FormInputPreview(id))
}
func sxFormInputDataIndex(c *gin.Context) {
	fid, _ := strconv.ParseUint(c.Query("form_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.FormInputDataList(uint(fid), page, 20)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}
func sxFormInputDataSave(c *gin.Context) { FormInputDataSubmitHandler(c) }
func sxFormInputDataDelete(c *gin.Context) {
	id := getID(c)
	service.FormInputApiDelete(id, c.GetUint("user_id"))
	response.OK(c, nil)
}

func sxFormInputVerifySend(c *gin.Context) {
	account := c.PostForm("accounts")
	if account == "" {
		account = c.PostForm("mobile")
	}
	service.SendVerifyCode(account, "forminput")
	response.OK(c, nil)
}

// ===== 辅助函数 =====

func getID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if id == 0 {
		id, _ = strconv.ParseUint(c.PostForm("id"), 10, 64)
	}
	if id == 0 {
		id, _ = strconv.ParseUint(c.Query("goods_id"), 10, 64)
	}
	if id == 0 {
		id, _ = strconv.ParseUint(c.PostForm("order_id"), 10, 64)
	}
	return uint(id)
}

func parseUintSlice(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var ids []uint
	for _, p := range parts {
		id, _ := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}
