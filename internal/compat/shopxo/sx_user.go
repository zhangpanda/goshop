package shopxo

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func sxCommon(c *gin.Context) {
	// 构造 base/common 兼容返回结构（字段形状参考 shopxo-uniapp 常见期望）
	config := map[string]interface{}{
		"common_site_type":                           service.GetConfig("common_site_type"),
		"common_shop_notice":                         service.GetConfig("common_shop_notice"),
		"common_app_is_enable_search":                1,
		"common_app_is_online_service":               0,
		"common_app_customer_service_tel":            service.CustomerServiceTel(),
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

