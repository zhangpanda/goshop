package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zhangpanda/goshop/global"
	"gorm.io/gorm"
	"github.com/zhangpanda/goshop/internal/model"
	auth_pkg "github.com/zhangpanda/goshop/pkg/auth"
)

// ==================== 3. 多平台OAuth登录绑定 ====================

type PlatformLoginReq struct {
	Platform string `json:"platform" binding:"required"` // alipay,baidu,toutiao,qq,kuaishou
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type PlatformLoginResp struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
	IsNew bool       `json:"is_new"`
}

// 各平台code换token的配置
type platformOAuthConfig struct {
	TokenURL     string
	UserInfoURL  string
	AppIDKey     string
	SecretKey    string
	CodeKey      string
	OpenIDField  string
	UnionIDField string
}

var platformConfigs = map[string]platformOAuthConfig{
	"alipay": {
		TokenURL:    "https://openapi.alipay.com/gateway.do?method=alipay.system.oauth.token&grant_type=authorization_code",
		OpenIDField: "user_id",
	},
	"baidu": {
		TokenURL:    "https://openapi.baidu.com/oauth/2.0/token?grant_type=authorization_code",
		OpenIDField: "openid",
	},
	"toutiao": {
		TokenURL:    "https://developer.toutiao.com/api/apps/jscode2session",
		OpenIDField: "openid",
	},
	"qq": {
		TokenURL:    "https://api.q.qq.com/sns/jscode2session",
		OpenIDField: "openid",
	},
	"kuaishou": {
		TokenURL:    "https://open.kuaishou.com/oauth2/mp/code2session",
		OpenIDField: "open_id",
	},
}

func PlatformLogin(req *PlatformLoginReq) (*PlatformLoginResp, error) {
	// 从小程序配置中获取appid/secret
	var mini model.AppMini
	global.DB.Where("platform = ? AND status = 1", req.Platform).First(&mini)
	if mini.ID == 0 {
		return nil, fmt.Errorf("平台 %s 未配置", req.Platform)
	}

	cfg, ok := platformConfigs[req.Platform]
	if !ok {
		return nil, fmt.Errorf("不支持的平台: %s", req.Platform)
	}

	// code换openid
	openID, unionID, err := exchangeCode(cfg, mini.AppID, mini.AppSecret, req.Code)
	if err != nil {
		return nil, fmt.Errorf("%s登录失败: %w", req.Platform, err)
	}

	// 查找或创建用户
	var platform model.UserPlatform
	global.DB.Where("platform = ? AND openid = ?", req.Platform, openID).First(&platform)

	var user model.User
	isNew := false

	if platform.ID > 0 {
		global.DB.First(&user, platform.UserID)
	} else {
		// 尝试通过unionid关联
		if unionID != "" {
			global.DB.Where("platform = ? AND unionid = ?", req.Platform, unionID).First(&platform)
			if platform.ID > 0 {
				global.DB.First(&user, platform.UserID)
			}
		}
		if user.ID == 0 {
			// 新用户
			user = model.User{Nickname: req.Nickname, Avatar: req.Avatar, Status: 1}
			global.DB.Create(&user)
			isNew = true
		}
		// 绑定平台
		global.DB.Create(&model.UserPlatform{
			UserID: user.ID, Platform: req.Platform, OpenID: openID, UnionID: unionID,
		})
	}

	// 更新昵称头像
	if req.Nickname != "" || req.Avatar != "" {
		updates := map[string]interface{}{}
		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.Avatar != "" {
			updates["avatar"] = req.Avatar
		}
		global.DB.Model(&user).Updates(updates)
	}

	token, _ := generateUserToken(user.ID)
	return &PlatformLoginResp{Token: token, User: user, IsNew: isNew}, nil
}

func exchangeCode(cfg platformOAuthConfig, appID, secret, code string) (openID, unionID string, err error) {
	reqURL := fmt.Sprintf("%s&appid=%s&secret=%s&js_code=%s", cfg.TokenURL, appID, secret, code)
	resp, err := http.Get(reqURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if v, ok := result[cfg.OpenIDField]; ok {
		openID = fmt.Sprint(v)
	}
	if v, ok := result["unionid"]; ok {
		unionID = fmt.Sprint(v)
	}
	if openID == "" {
		return "", "", fmt.Errorf("获取openid失败: %v", result)
	}
	return openID, unionID, nil
}

func generateUserToken(userID uint) (string, error) {
	return auth_pkg.GenerateToken(userID, false, global.Cfg.JWT.Secret, global.Cfg.JWT.Expire)
}

// ==================== 5. 积分锁定释放完整逻辑 ====================

// OrderGoodsIntegralGiving 订单完成后赠送积分（先锁定）
func OrderGoodsIntegralGiving(userID, orderID uint, payAmount int64) error {
	points := int(payAmount / 100)
	if points <= 0 {
		return nil
	}
	tx := global.DB.Begin()
	// 增加锁定积分
	tx.Model(&model.User{}).Where("id = ?", userID).Update("locking_integral", gorm.Expr("locking_integral + ?", points))
	// 记录赠送日志（状态0=锁定中）
	tx.Create(&model.GoodsGiveIntegralLog{
		OrderID: orderID, UserID: userID, Integral: points, Status: 0,
	})
	return tx.Commit().Error
}

// CronIntegralRelease 定时释放锁定积分（赠送后N天释放）
func CronIntegralRelease(limitMinutes int) (sucs, fail int) {
	if limitMinutes <= 0 {
		limitMinutes = 21600 // 默认15天=21600分钟
	}
	deadline := time.Now().Add(-time.Duration(limitMinutes) * time.Minute)
	var logs []model.GoodsGiveIntegralLog
	global.DB.Where("status = 0 AND created_at < ?", deadline).Limit(200).Find(&logs)
	for _, log := range logs {
		tx := global.DB.Begin()
		// 锁定积分转正式积分
		tx.Model(&model.User{}).Where("id = ?", log.UserID).
			Update("locking_integral", gorm.Expr("locking_integral - ?", log.Integral))
		tx.Model(&model.User{}).Where("id = ?", log.UserID).
			Update("points", gorm.Expr("points + ?", log.Integral))
		tx.Model(&log).Updates(map[string]interface{}{"status": 1, "updated_at": time.Now()})
		// 积分日志
		tx.Create(&model.PointsLog{
			UserID: log.UserID, Points: log.Integral, Type: "goods_integral",
			RefID: log.OrderID, Remark: fmt.Sprintf("商品赠送积分释放%d", log.Integral),
		})
		if err := tx.Commit().Error; err != nil {
			fail++
			continue
		}
		sucs++
	}
	return
}

// OrderGoodsIntegralRollback 售后退款扣减锁定积分
func OrderGoodsIntegralRollback(orderID, orderDetailID uint, refundAmount int64) error {
	var log model.GoodsGiveIntegralLog
	global.DB.Where("order_id = ? AND status = 0", orderID).First(&log)
	if log.ID == 0 {
		return nil
	}
	deduct := int(refundAmount / 100)
	if deduct <= 0 {
		return nil
	}
	if deduct > log.Integral {
		deduct = log.Integral
	}
	tx := global.DB.Begin()
	tx.Model(&model.User{}).Where("id = ?", log.UserID).
		Update("locking_integral", gorm.Expr("GREATEST(locking_integral - ?, 0)", deduct))
	tx.Model(&log).Update("integral", log.Integral-deduct)
	if log.Integral-deduct <= 0 {
		tx.Model(&log).Update("status", 2) // 关闭
	}
	return tx.Commit().Error
}

// ==================== 7. 微信小程序发货信息录入 ====================

func OrderDeliverySyncWeixin(orderNo, tradeNo, openID, goodsTitle, expressName, expressNumber, tel string, orderModel int8) error {
	cfg := global.Cfg.Wechat
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil
	}
	// 获取access_token
	tokenURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", cfg.AppID, cfg.AppSecret)
	tokenResp, err := http.Get(tokenURL)
	if err != nil {
		return err
	}
	defer tokenResp.Body.Close()
	var tokenResult struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(tokenResp.Body).Decode(&tokenResult)
	if tokenResult.AccessToken == "" {
		return fmt.Errorf("获取access_token失败")
	}

	// 订单模式映射微信物流类型
	logisticsType := 1 // 默认实体物流
	switch orderModel {
	case 1:
		logisticsType = 2 // 同城配送
	case 2:
		logisticsType = 4 // 用户自提
	case 3:
		logisticsType = 3 // 虚拟商品
	}

	body := map[string]interface{}{
		"order_key": map[string]interface{}{
			"order_number_type": 2,
			"transaction_id":    tradeNo,
		},
		"logistics_type": logisticsType,
		"delivery_mode":  1,
		"upload_time":    time.Now().Format("2006-01-02T15:04:05.000+08:00"),
		"shipping_list": []map[string]interface{}{{
			"tracking_no":     expressNumber,
			"express_company": expressName,
			"item_desc":       goodsTitle,
			"contact":         map[string]string{"receiver_contact": tel},
		}},
		"payer": map[string]string{"openid": openID},
	}
	bodyJSON, _ := json.Marshal(body)
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/sec/order/upload_shipping_info?access_token=%s", tokenResult.AccessToken)
	http.Post(url, "application/json", bytes.NewReader(bodyJSON))
	return nil
}

// ==================== 8. 短信/邮件日志完整service ====================

func SmsLogAdd(phone, content, typ string, status int8) {
	global.DB.Create(&model.SmsLog{Phone: phone, Content: content, Type: typ, Status: status})
}

func SmsLogList(page, pageSize int) ([]model.SmsLog, int64, error) {
	var total int64
	global.DB.Model(&model.SmsLog{}).Count(&total)
	var list []model.SmsLog
	err := global.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func SmsLogDelete(ids []uint) error {
	return global.DB.Where("id IN ?", ids).Delete(&model.SmsLog{}).Error
}

func SmsLogAllDelete() error {
	return global.DB.Where("1=1").Delete(&model.SmsLog{}).Error
}

func EmailLogAdd(email, title, content string, status int8) {
	global.DB.Create(&model.EmailLog{Email: email, Title: title, Content: content, Status: status})
}

func EmailLogList(page, pageSize int) ([]model.EmailLog, int64, error) {
	var total int64
	global.DB.Model(&model.EmailLog{}).Count(&total)
	var list []model.EmailLog
	err := global.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func EmailLogDelete(ids []uint) error {
	return global.DB.Where("id IN ?", ids).Delete(&model.EmailLog{}).Error
}

func EmailLogAllDelete() error {
	return global.DB.Where("1=1").Delete(&model.EmailLog{}).Error
}

// ==================== 9. 商品缺失方法 ====================

// GoodsStock 查询商品库存（按规格）
func GoodsStock(goodsID uint, spec string) (int, error) {
	if spec == "" {
		var total int64
		global.DB.Model(&model.GoodsSKU{}).Where("goods_id = ? AND status = 1", goodsID).
			Select("COALESCE(SUM(stock),0)").Scan(&total)
		return int(total), nil
	}
	var sku model.GoodsSKU
	global.DB.Where("goods_id = ? AND specs LIKE ? AND status = 1", goodsID, "%"+spec+"%").First(&sku)
	return sku.Stock, nil
}

// GoodsSpecDetail 获取规格详情（选中规格后返回价格/库存/图片）
type GoodsSpecDetailResp struct {
	SKUID   uint    `json:"sku_id"`
	Price   int64   `json:"price"`
	Stock   int     `json:"stock"`
	Image   string  `json:"image"`
	Coding  string  `json:"coding"`
	Barcode string  `json:"barcode"`
	Weight  float64 `json:"weight"`
}

func GoodsSpecDetail(goodsID uint, specValues string) (*GoodsSpecDetailResp, error) {
	// 先查GoodsSpecBase
	var base model.GoodsSpecBase
	global.DB.Where("goods_id = ? AND spec_values = ?", goodsID, specValues).First(&base)
	if base.ID > 0 {
		return &GoodsSpecDetailResp{
			Price: base.Price, Stock: base.Inventory, Coding: base.Coding,
			Barcode: base.Barcode, Weight: base.Weight,
		}, nil
	}
	// 回退查SKU
	var sku model.GoodsSKU
	global.DB.Where("goods_id = ? AND name = ?", goodsID, specValues).First(&sku)
	if sku.ID == 0 {
		return nil, fmt.Errorf("规格不存在")
	}
	return &GoodsSpecDetailResp{SKUID: sku.ID, Price: sku.Price, Stock: sku.Stock, Image: sku.Image}, nil
}

// GoodsScore 商品评分统计
type GoodsScoreData struct {
	Total   int64   `json:"total"`
	Average float64 `json:"average"`
	Star5   int64   `json:"star5"`
	Star4   int64   `json:"star4"`
	Star3   int64   `json:"star3"`
	Star2   int64   `json:"star2"`
	Star1   int64   `json:"star1"`
}

func GoodsScore(goodsID uint) *GoodsScoreData {
	d := &GoodsScoreData{}
	global.DB.Model(&model.Review{}).Where("goods_id = ?", goodsID).Count(&d.Total)
	if d.Total == 0 {
		return d
	}
	global.DB.Model(&model.Review{}).Where("goods_id = ?", goodsID).Select("COALESCE(AVG(rating),0)").Scan(&d.Average)
	for i := 1; i <= 5; i++ {
		var c int64
		global.DB.Model(&model.Review{}).Where("goods_id = ? AND rating = ?", goodsID, i).Count(&c)
		switch i {
		case 1:
			d.Star1 = c
		case 2:
			d.Star2 = c
		case 3:
			d.Star3 = c
		case 4:
			d.Star4 = c
		case 5:
			d.Star5 = c
		}
	}
	return d
}

// HomeFloorList 首页楼层商品（按一级分类分组）
type HomeFloor struct {
	CategoryID   uint          `json:"category_id"`
	CategoryName string        `json:"category_name"`
	Goods        []model.Goods `json:"goods"`
}

func HomeFloorList(maxCount int) []HomeFloor {
	if maxCount <= 0 {
		maxCount = 8
	}
	var cats []model.Category
	global.DB.Where("parent_id = 0 AND status = 1").Order("sort DESC").Find(&cats)
	var floors []HomeFloor
	for _, cat := range cats {
		var goods []model.Goods
		global.DB.Where("category_id = ? AND status = 1", cat.ID).
			Preload("SKUs").Order("sort DESC, sales_count DESC").Limit(maxCount).Find(&goods)
		if len(goods) > 0 {
			floors = append(floors, HomeFloor{CategoryID: cat.ID, CategoryName: cat.Name, Goods: goods})
		}
	}
	return floors
}

// GuessYouLike 猜你喜欢（基于浏览记录+销量）
func GuessYouLike(userID uint, limit int) []model.Goods {
	if limit <= 0 {
		limit = 10
	}
	// 取用户最近浏览的分类
	var catIDs []uint
	if userID > 0 {
		global.DB.Model(&model.BrowseHistory{}).
			Joins("JOIN goods ON goods.id = browse_histories.goods_id").
			Where("browse_histories.user_id = ?", userID).
			Select("DISTINCT goods.category_id").Limit(5).Pluck("goods.category_id", &catIDs)
	}
	db := global.DB.Where("status = 1")
	if len(catIDs) > 0 {
		db = db.Where("category_id IN ?", catIDs)
	}
	var list []model.Goods
	db.Preload("SKUs").Order("sales_count DESC, id DESC").Limit(limit).Find(&list)
	return list
}

// ==================== 10. 订单缺失方法 ====================

// OrderStatusGroupTotal 订单状态分组统计
func OrderStatusGroupTotal(userID uint) map[string]int64 {
	result := map[string]int64{}
	type StatusCount struct {
		Status int8
		Count  int64
	}
	var counts []StatusCount
	db := global.DB.Model(&model.Order{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	db.Select("status, COUNT(*) as count").Group("status").Find(&counts)
	names := map[int8]string{0: "pending", 1: "paid", 2: "shipped", 3: "completed", 4: "cancelled", 5: "refunded", 6: "booking"}
	for _, c := range counts {
		if name, ok := names[c.Status]; ok {
			result[name] = c.Count
		}
	}
	// 售后数量
	var asCount int64
	db2 := global.DB.Model(&model.OrderAftersale{})
	if userID > 0 {
		db2 = db2.Where("user_id = ?", userID)
	}
	db2.Where("status IN ?", []int8{0, 1, 2}).Count(&asCount)
	result["aftersale"] = asCount
	return result
}

// OrderOperateData 订单可执行的操作按钮
type OrderOperate struct {
	CanCancel    bool `json:"can_cancel"`
	CanPay       bool `json:"can_pay"`
	CanReceive   bool `json:"can_receive"`
	CanDelete    bool `json:"can_delete"`
	CanReview    bool `json:"can_review"`
	CanAftersale bool `json:"can_aftersale"`
	CanRefund    bool `json:"can_refund"`
}

func OrderOperateButtons(order *model.Order) *OrderOperate {
	op := &OrderOperate{}
	switch order.Status {
	case model.OrderStatusBooking:
		op.CanCancel = true
	case model.OrderStatusPending:
		op.CanCancel = true
		op.CanPay = true
	case model.OrderStatusPaid:
		op.CanRefund = true
		op.CanAftersale = true
	case model.OrderStatusShipped:
		op.CanReceive = true
		op.CanAftersale = true
	case model.OrderStatusCompleted:
		op.CanReview = true
		op.CanAftersale = true
		op.CanDelete = true
	case model.OrderStatusCancelled:
		op.CanDelete = true
	}
	return op
}

// AdminOrderPayUnderLine 管理员确认线下支付收款
func AdminOrderPayUnderLine(orderID, adminID uint) error {
	var order model.Order
	if err := global.DB.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("订单不存在")
	}
	if order.Status != model.OrderStatusPending {
		return fmt.Errorf("订单状态不允许确认收款")
	}
	now := time.Now()
	global.DB.Model(&order).Updates(map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": &now})
	AddOrderStatusHistory(orderID, model.OrderStatusPending, model.OrderStatusPaid,
		fmt.Sprintf("管理员(ID:%d)确认线下收款", adminID), "管理员")
	NotifyOrderStatus(order.UserID, orderID, order.OrderNo, "paid")
	return nil
}
