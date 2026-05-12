package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// ==================== 1. 订单拆单 ====================

func SplitOrderByWarehouse(userID uint, req *CreateOrderReq) ([]*model.Order, error) {
	var carts []model.Cart
	if err := app.Must().DB.Where("id IN ? AND user_id = ?", req.CartIDs, userID).
		Preload("Goods").Preload("SKU").Find(&carts).Error; err != nil || len(carts) == 0 {
		return nil, errors.New("购物车为空")
	}
	// 按仓库+SKU级别分组：查WarehouseGoodsSpec精确到规格
	type groupKey struct{ WarehouseID uint }
	groups := map[groupKey][]model.Cart{}
	for _, c := range carts {
		var ws model.WarehouseGoodsSpec
		// WHERE 里明确写 warehouse_goods_specs.goods_id / sku_id：多表 JOIN 中
		// warehouse_goods 也有这两列，裸写 goods_id/sku_id 在 MySQL 下 1052 ambiguous。
		app.Must().DB.Where("warehouse_goods_specs.goods_id = ? AND warehouse_goods_specs.sku_id = ? AND warehouse_goods_specs.inventory > 0", c.GoodsID, c.SKUID).
			Joins("JOIN warehouse_goods ON warehouse_goods.warehouse_id = warehouse_goods_specs.warehouse_id AND warehouse_goods.goods_id = warehouse_goods_specs.goods_id AND warehouse_goods.is_enable = 1").
			Joins("JOIN warehouses ON warehouses.id = warehouse_goods_specs.warehouse_id AND warehouses.is_enable = 1").
			Order("warehouses.level DESC").First(&ws)
		key := groupKey{WarehouseID: ws.WarehouseID}
		groups[key] = append(groups[key], c)
	}
	if len(groups) <= 1 {
		return nil, nil // 不需要拆单
	}
	var orders []*model.Order
	for _, groupCarts := range groups {
		ids := make([]uint, len(groupCarts))
		for i, c := range groupCarts {
			ids[i] = c.ID
		}
		subReq := *req
		subReq.CartIDs = ids
		order, err := CreateOrder(userID, &subReq)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// ==================== 2. 完整统计 ====================

type StatisticalData struct {
	Today     StatItem          `json:"today"`
	Yesterday StatItem          `json:"yesterday"`
	Week      StatItem          `json:"week"`
	Month     StatItem          `json:"month"`
	Trend     []DayStat         `json:"trend"`
	GoodsTop  []GoodsRank       `json:"goods_top"`
	UserTop   []UserRank        `json:"user_top"`
	OrderDist []OrderStatusDist `json:"order_dist"`
	// 待处理事项
	OrderPendingCount     int64 `json:"order_pending_count"`
	AftersalePendingCount int64 `json:"aftersale_pending_count"`
	GoodsOfflineCount     int64 `json:"goods_offline_count"`
	ReviewPendingCount    int64 `json:"review_pending_count"`
	// 支付方式统计
	PayTypeStats []PayTypeStat `json:"pay_type_stats"`
	// 地域分布
	RegionStats []RegionStat `json:"region_stats"`
	// 新增用户趋势
	NewUserTrend []DayStat `json:"new_user_trend"`
}
type StatItem struct {
	OrderCount int64 `json:"order_count"`
	Sales      int64 `json:"sales"`
	UserCount  int64 `json:"user_count"`
	GoodsCount int64 `json:"goods_count"`
}
type DayStat struct {
	Date  string `json:"date"`
	Sales int64  `json:"sales"`
	Count int64  `json:"count"`
}
type GoodsRank struct {
	GoodsID uint   `json:"goods_id"`
	Title   string `json:"title"`
	Sales   int64  `json:"sales"`
}
type UserRank struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Amount   int64  `json:"amount"`
}

// userDisplayName 仪表盘等场景展示用名称：优先昵称，其次用户名、手机号，皆空则「用户ID」。
func userDisplayName(u *model.User, id uint) string {
	if u == nil {
		return fmt.Sprintf("用户%d", id)
	}
	if s := strings.TrimSpace(u.Nickname); s != "" {
		return s
	}
	if s := strings.TrimSpace(u.Username); s != "" {
		return s
	}
	if s := strings.TrimSpace(u.Phone); s != "" {
		return s
	}
	return fmt.Sprintf("用户%d", id)
}

type OrderStatusDist struct {
	Status int8  `json:"status"`
	Count  int64 `json:"count"`
}

func GetStatistical(days int) *StatisticalData {
	if days <= 0 {
		days = 30
	}
	data := &StatisticalData{}
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	fillStat := func(date string, d int) StatItem {
		var s StatItem
		q := app.Must().DB.Model(&model.Order{}).Where("status > 0")
		if d == 1 {
			q = q.Where("DATE(created_at) = ?", date)
		} else {
			q = q.Where("created_at >= ?", now.AddDate(0, 0, -d))
		}
		q.Count(&s.OrderCount)
		q.Select("COALESCE(SUM(pay_amount),0)").Scan(&s.Sales)
		if d == 1 {
			app.Must().DB.Model(&model.User{}).Where("DATE(created_at) = ?", date).Count(&s.UserCount)
		} else {
			app.Must().DB.Model(&model.User{}).Where("created_at >= ?", now.AddDate(0, 0, -d)).Count(&s.UserCount)
		}
		app.Must().DB.Model(&model.Goods{}).Where("status = 1").Count(&s.GoodsCount)
		return s
	}
	data.Today = fillStat(today, 1)
	data.Yesterday = fillStat(yesterday, 1)
	data.Week = fillStat("", 7)
	data.Month = fillStat("", 30)

	// 趋势
	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		var ds DayStat
		ds.Date = d
		app.Must().DB.Model(&model.Order{}).Where("DATE(created_at) = ? AND status > 0", d).
			Select("COALESCE(SUM(pay_amount),0) as sales, COUNT(*) as count").Scan(&ds)
		ds.Date = d
		data.Trend = append(data.Trend, ds)
	}
	// 商品销量Top10
	app.Must().DB.Model(&model.Goods{}).Select("id as goods_id, title, sales_count as sales").
		Order("sales_count DESC").Limit(10).Find(&data.GoodsTop)
	// 用户消费Top10
	app.Must().DB.Model(&model.Order{}).Select("user_id, SUM(pay_amount) as amount").
		Where("status > 0").Group("user_id").Order("amount DESC").Limit(10).Find(&data.UserTop)
	for i := range data.UserTop {
		var u model.User
		if err := app.Must().DB.Select("nickname", "username", "phone").First(&u, data.UserTop[i].UserID).Error; err != nil {
			data.UserTop[i].Nickname = fmt.Sprintf("用户%d", data.UserTop[i].UserID)
			continue
		}
		data.UserTop[i].Nickname = userDisplayName(&u, data.UserTop[i].UserID)
	}
	// 订单状态分布
	app.Must().DB.Model(&model.Order{}).Select("status, COUNT(*) as count").Group("status").Find(&data.OrderDist)

	// 待处理事项
	app.Must().DB.Model(&model.Order{}).Where("status = 0").Count(&data.OrderPendingCount)
	app.Must().DB.Model(&model.OrderAftersale{}).Where("status = 0").Count(&data.AftersalePendingCount)
	app.Must().DB.Model(&model.Goods{}).Where("status = 0").Count(&data.GoodsOfflineCount)
	app.Must().DB.Model(&model.Review{}).Where("reply = '' OR reply IS NULL").Count(&data.ReviewPendingCount)

	// 支付方式统计
	app.Must().DB.Model(&model.PayLog{}).Select("client_type, COUNT(*) as count, COALESCE(SUM(total_price),0) as amount").
		Where("status = 1").Group("client_type").Find(&data.PayTypeStats)

	// 地域分布Top10（从订单地址JSON提取省份）
	app.Must().DB.Raw(`SELECT JSON_UNQUOTE(JSON_EXTRACT(address, '$.province')) as province, COUNT(*) as count 
		FROM orders WHERE status > 0 AND address != '' AND address IS NOT NULL 
		GROUP BY province HAVING province IS NOT NULL AND province != '' 
		ORDER BY count DESC LIMIT 10`).Find(&data.RegionStats)

	// 新增用户趋势
	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		var ds DayStat
		ds.Date = d
		app.Must().DB.Model(&model.User{}).Where("DATE(created_at) = ?", d).Count(&ds.Count)
		data.NewUserTrend = append(data.NewUserTrend, ds)
	}

	return data
}

// ==================== 3. DiyApi数据源 ====================

type DiyApiParams struct {
	CategoryID uint   `form:"category_id" json:"category_id"`
	BrandID    uint   `form:"brand_id" json:"brand_id"`
	OrderBy    string `form:"order_by" json:"order_by"` // sales, new, price
	Limit      int    `form:"limit" json:"limit"`
}

func DiyApiGoodsAutoData(p *DiyApiParams) ([]model.Goods, error) {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	db := app.Must().DB.Where("status = 1")
	if p.CategoryID > 0 {
		db = db.Where("category_id = ?", p.CategoryID)
	}
	if p.BrandID > 0 {
		db = db.Where("brand_id = ?", p.BrandID)
	}
	order := "sort DESC, id DESC"
	switch p.OrderBy {
	case "sales":
		order = "sales_count DESC"
	case "new":
		order = "id DESC"
	}
	var list []model.Goods
	err := db.Preload("SKUs").Order(order).Limit(p.Limit).Find(&list).Error
	return list, err
}

func DiyApiArticleAutoData(categoryID uint, limit int) ([]model.Article, error) {
	if limit <= 0 {
		limit = 10
	}
	db := app.Must().DB.Where("status = 1")
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	var list []model.Article
	return list, db.Order("sort DESC, id DESC").Limit(limit).Find(&list).Error
}

func DiyApiBrandAutoData(limit int) ([]model.Brand, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []model.Brand
	return list, app.Must().DB.Where("status = 1").Order("sort DESC").Limit(limit).Find(&list).Error
}

func DiyApiGoodsFavorAutoData(userID uint, limit int) ([]model.Favorite, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.Favorite
	return list, app.Must().DB.Where("user_id = ?", userID).Preload("Goods").Order("id DESC").Limit(limit).Find(&list).Error
}

func DiyApiGoodsBrowseAutoData(userID uint, limit int) ([]model.BrowseHistory, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.BrowseHistory
	return list, app.Must().DB.Where("user_id = ?", userID).Preload("Goods").Order("updated_at DESC").Limit(limit).Find(&list).Error
}

// ==================== 4. 小程序管理 ====================

type AppMiniReq struct {
	Platform  string `json:"platform" binding:"required"` // weixin, alipay, baidu, toutiao, qq, kuaishou
	Title     string `json:"title"`
	Describe  string `json:"describe"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	Status    int8   `json:"status"`
}

func SaveAppMini(req *AppMiniReq) error {
	var m model.AppMini
	app.Must().DB.Where("platform = ?", req.Platform).First(&m)
	if m.ID > 0 {
		return app.Must().DB.Model(&m).Updates(map[string]interface{}{
			"title": req.Title, "describe": req.Describe,
			"app_id": req.AppID, "app_secret": req.AppSecret, "status": req.Status,
		}).Error
	}
	return app.Must().DB.Create(&model.AppMini{
		Platform: req.Platform, Title: req.Title, Describe: req.Describe,
		AppID: req.AppID, AppSecret: req.AppSecret, Status: req.Status,
	}).Error
}

func GetAppMiniList() ([]model.AppMini, error) {
	var list []model.AppMini
	return list, app.Must().DB.Order("id ASC").Find(&list).Error
}

func DeleteAppMini(id uint) error { return app.Must().DB.Delete(&model.AppMini{}, id).Error }

// ==================== 5. 站点配置 ====================

func GetSiteConfig() map[string]string {
	var configs []model.Config
	app.Must().DB.Where("`group` IN ('site','base','seo','app') OR `key` LIKE 'home_site%' OR `key` LIKE 'home_seo%' OR `key` LIKE 'home_footer%'").Find(&configs)
	result := make(map[string]string, len(configs))
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result
}

func SaveSiteConfig(configs map[string]string) {
	for k, v := range configs {
		SetConfig(k, v, "site", "")
	}
}

func GetSelfExtractionAddressList() []map[string]interface{} {
	raw := GetConfig("site_self_extraction_address")
	if raw == "" {
		return nil
	}
	var list []map[string]interface{}
	json.Unmarshal([]byte(raw), &list)
	return list
}

func SaveSelfExtractionAddress(list []map[string]interface{}) {
	data, _ := json.Marshal(list)
	SetConfig("site_self_extraction_address", string(data), "site", "自提点地址")
}

// ==================== 6. 动态表格 ====================

type FormTableParams struct {
	Table         string           `json:"table" binding:"required"`
	Keyword       string           `json:"keyword"`
	KeywordFields string           `json:"keyword_fields"` // 逗号分隔
	Where         []FormTableWhere `json:"where"`
	OrderBy       string           `json:"order_by"`
	Page          int              `json:"page"`
	PageSize      int              `json:"page_size"`
}
type FormTableWhere struct {
	Field string      `json:"field"`
	Op    string      `json:"op"` // =, !=, >, <, >=, <=, like, in
	Value interface{} `json:"value"`
}

// 允许查询的表白名单（不含 users/admins 等敏感表）
var allowedTables = map[string]bool{
	"orders": true, "goods": true, "order_aftersales": true,
	"reviews": true, "coupons": true, "brands": true, "articles": true,
	"pay_logs": true, "refund_logs": true, "messages": true, "error_logs": true,
	"warehouses": true, "plugins": true, "answers": true, "sms_logs": true,
	"email_logs": true, "search_histories": true, "attachments": true,
}

var allowedOps = map[string]bool{"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true, "like": true, "in": true}
var safeFieldRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
var safeOrderRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\s+(ASC|DESC|asc|desc)$`)

func FormTableQuery(p *FormTableParams) (int64, []map[string]interface{}, error) {
	if !allowedTables[p.Table] {
		return 0, nil, fmt.Errorf("不允许查询表: %s", p.Table)
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	db := app.Must().DB.Table(p.Table)
	if p.Keyword != "" && p.KeywordFields != "" {
		fields := strings.Split(p.KeywordFields, ",")
		var conds []string
		var args []interface{}
		for _, f := range fields {
			if safeFieldRe.MatchString(f) {
				conds = append(conds, f+" LIKE ?")
				args = append(args, "%"+p.Keyword+"%")
			}
		}
		if len(conds) > 0 {
			db = db.Where(strings.Join(conds, " OR "), args...)
		}
	}
	for _, w := range p.Where {
		if !safeFieldRe.MatchString(w.Field) || !allowedOps[strings.ToLower(w.Op)] {
			continue
		}
		switch strings.ToLower(w.Op) {
		case "like":
			db = db.Where(w.Field+" LIKE ?", "%"+fmt.Sprint(w.Value)+"%")
		case "in":
			db = db.Where(w.Field+" IN ?", w.Value)
		default:
			db = db.Where(w.Field+" "+w.Op+" ?", w.Value)
		}
	}
	var total int64
	db.Count(&total)
	if p.OrderBy != "" && safeOrderRe.MatchString(strings.TrimSpace(p.OrderBy)) {
		db = db.Order(p.OrderBy)
	} else {
		db = db.Order("id DESC")
	}
	var results []map[string]interface{}
	err := db.Offset((p.Page - 1) * p.PageSize).Limit(p.PageSize).Find(&results).Error
	return total, results, err
}

// ==================== 7. 二维码 ====================

func GenerateQRCodeURL(content string) string {
	// 使用Google Chart API生成二维码URL，前端直接用<img>展示
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=%s", content)
}

// ==================== 8. SQL控制台 ====================

var forbiddenSQLKeywords = []string{
	"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "TRUNCATE", "CREATE",
	"GRANT", "REVOKE", "INTO OUTFILE", "INTO DUMPFILE", "LOAD_FILE",
	"SLEEP", "BENCHMARK", "EXEC", "EXECUTE",
}

func ExecuteSQL(sqlStr string) ([]map[string]interface{}, error) {
	trimmed := strings.TrimSpace(sqlStr)
	upper := strings.ToUpper(trimmed)
	// 禁止多语句
	if strings.Contains(trimmed, ";") && strings.TrimRight(trimmed, "; \t\n") != strings.TrimRight(strings.SplitN(trimmed, ";", 2)[0], " \t\n") {
		return nil, errors.New("禁止多语句执行")
	}
	// 仅允许 SELECT/SHOW/DESC/EXPLAIN 开头
	allowed := false
	for _, prefix := range []string{"SELECT", "SHOW", "DESC", "EXPLAIN"} {
		if strings.HasPrefix(upper, prefix) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("仅允许 SELECT/SHOW/DESC/EXPLAIN 查询")
	}
	// 检查危险关键字（全文扫描，防注释绕过）
	for _, kw := range forbiddenSQLKeywords {
		if strings.Contains(upper, kw) {
			return nil, fmt.Errorf("SQL 包含禁止关键字: %s", kw)
		}
	}
	// 禁止访问系统库/元数据（开启 sql_console 时纵深防御）
	for _, blocked := range []string{
		"INFORMATION_SCHEMA", "PERFORMANCE_SCHEMA", "MYSQL.", "SYS.",
	} {
		if strings.Contains(upper, blocked) {
			return nil, errors.New("禁止查询系统库或元数据")
		}
	}
	// 带超时执行
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var results []map[string]interface{}
	err := app.Must().DB.WithContext(ctx).Raw(trimmed).Limit(1000).Find(&results).Error
	return results, err
}

// ==================== 9. 系统信息 ====================

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	AppVersion   string `json:"app_version"`
	DBVersion    string `json:"db_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	StartTime    string `json:"start_time"`
}

var appStartTime = time.Now()

func GetSystemInfo() *SystemInfo {
	var dbVer string
	app.Must().DB.Raw("SELECT VERSION()").Scan(&dbVer)
	return &SystemInfo{
		GoVersion:    runtime.Version(),
		AppVersion:   "1.0.0",
		DBVersion:    dbVer,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		StartTime:    appStartTime.Format(time.DateTime),
	}
}
