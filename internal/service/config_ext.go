package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"gorm.io/gorm"
)

// ========== 多语言 ==========

type MultilingualConfig struct {
	DefaultLang string   `json:"default_lang"` // zh, en, cht
	Available   []string `json:"available"`
}

func GetMultilingualConfig() *MultilingualConfig {
	defaultLang := GetConfig("multilingual_default")
	if defaultLang == "" {
		defaultLang = "zh"
	}
	available := GetConfig("multilingual_available")
	if available == "" {
		available = "zh,en,cht"
	}
	return &MultilingualConfig{DefaultLang: defaultLang, Available: strings.Split(available, ",")}
}

func SetMultilingualConfig(defaultLang string, available []string) {
	SetConfig("multilingual_default", defaultLang, "multilingual", "默认语言")
	SetConfig("multilingual_available", strings.Join(available, ","), "multilingual", "可用语言")
}

// GetLangPack 获取语言包（从config表读取，key格式: lang_{code}_{module}_{key}）
func GetLangPack(lang, module string) (map[string]string, error) {
	var configs []model.Config
	prefix := fmt.Sprintf("lang_%s_%s_", lang, module)
	global.DB.Where("`key` LIKE ?", prefix+"%").Find(&configs)
	result := make(map[string]string, len(configs))
	for _, c := range configs {
		k := strings.TrimPrefix(c.Key, prefix)
		result[k] = c.Value
	}
	return result, nil
}

// ========== 货币配置 ==========

type CurrencyConfig struct {
	Symbol string  `json:"symbol"`
	Code   string  `json:"code"`
	Rate   float64 `json:"rate"`
	Name   string  `json:"name"`
}

func GetCurrencyConfig() *CurrencyConfig {
	symbol := GetConfig("currency_symbol")
	if symbol == "" {
		symbol = "¥"
	}
	code := GetConfig("currency_code")
	if code == "" {
		code = "CNY"
	}
	name := GetConfig("currency_name")
	if name == "" {
		name = "人民币"
	}
	rate := 1.0
	if rs := GetConfig("currency_rate"); rs != "" {
		if v, err := strconv.ParseFloat(rs, 64); err == nil && v > 0 {
			rate = v
		}
	}
	return &CurrencyConfig{Symbol: symbol, Code: code, Rate: rate, Name: name}
}

func SetCurrencyConfig(cfg *CurrencyConfig) {
	SetConfig("currency_symbol", cfg.Symbol, "currency", "货币符号")
	SetConfig("currency_code", cfg.Code, "currency", "货币代码")
	SetConfig("currency_name", cfg.Name, "currency", "货币名称")
	SetConfig("currency_rate", fmt.Sprintf("%f", cfg.Rate), "currency", "汇率")
}

// SaveOrderCurrency 订单创建时保存货币信息
func SaveOrderCurrency(orderID uint, amount int64) {
	cfg := GetCurrencyConfig()
	global.DB.Create(&model.OrderCurrency{
		OrderID:  orderID,
		Currency: cfg.Code,
		Rate:     cfg.Rate,
		Amount:   amount,
	})
}

// ========== 订单预约模式 ==========

func IsBookingMode() bool {
	return GetConfig("order_is_booking") == "1"
}

// BookingConfirm 管理员确认预约订单（从预约状态变为待支付）
func BookingConfirm(orderID uint) error {
	var order model.Order
	if err := global.DB.First(&order, orderID).Error; err != nil {
		return fmt.Errorf("订单不存在")
	}
	if order.Status != model.OrderStatusBooking {
		return fmt.Errorf("订单状态不允许确认")
	}
	global.DB.Model(&order).Update("status", model.OrderStatusPending)
	AddOrderStatusHistory(orderID, model.OrderStatusBooking, model.OrderStatusPending, "管理员确认预约", "管理员")
	return nil
}

// ========== 数据导出Excel (CSV) ==========

type ExportReq struct {
	Type    string `json:"type" binding:"required"` // orders, users, goods
	Keyword string `json:"keyword"`
	Status  *int8  `json:"status"`
	IDs     []uint `json:"ids"` // 非空时仅导出这些 ID（与全表导出互斥）
}

func ExportData(w io.Writer, req *ExportReq) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	switch req.Type {
	case "orders":
		writer.Write([]string{"订单号", "用户ID", "总金额", "实付金额", "状态", "创建时间"})
		var list []model.Order
		db := global.DB.Model(&model.Order{})
		if len(req.IDs) > 0 {
			db = db.Where("id IN ?", req.IDs)
		} else if req.Status != nil {
			db = db.Where("status = ?", *req.Status)
		}
		db.Order("id DESC").Limit(10000).Find(&list)
		for _, o := range list {
			writer.Write([]string{o.OrderNo, fmt.Sprint(o.UserID), fmt.Sprint(o.TotalAmount), fmt.Sprint(o.PayAmount), fmt.Sprint(o.Status), o.CreatedAt.Format(time.DateTime)})
		}
	case "users":
		writer.Write([]string{"ID", "用户名", "昵称", "手机", "积分", "状态", "注册时间"})
		var list []model.User
		dbu := global.DB.Model(&model.User{})
		if len(req.IDs) > 0 {
			dbu = dbu.Where("id IN ?", req.IDs)
		}
		dbu.Order("id DESC").Limit(10000).Find(&list)
		for _, u := range list {
			writer.Write([]string{fmt.Sprint(u.ID), u.Username, u.Nickname, u.Phone, fmt.Sprint(u.Points), fmt.Sprint(u.Status), u.CreatedAt.Format(time.DateTime)})
		}
	case "goods":
		writer.Write([]string{"ID", "标题", "分类ID", "状态", "销量", "创建时间"})
		var list []model.Goods
		dbg := global.DB.Model(&model.Goods{})
		if len(req.IDs) > 0 {
			dbg = dbg.Where("id IN ?", req.IDs)
		}
		dbg.Order("id DESC").Limit(10000).Find(&list)
		for _, g := range list {
			writer.Write([]string{fmt.Sprint(g.ID), g.Title, fmt.Sprint(g.CategoryID), fmt.Sprint(g.Status), fmt.Sprint(g.SalesCount), g.CreatedAt.Format(time.DateTime)})
		}
	default:
		return fmt.Errorf("不支持的导出类型: %s", req.Type)
	}
	return nil
}

// ========== 账号注销 ==========

func UserLogout(userID uint) error {
	return RunInDBTx(global.DB, func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.Address{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.Cart{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.BrowseHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&model.Message{}).Error; err != nil {
			return err
		}
		return tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"status":   0,
			"username": fmt.Sprintf("deleted_%d_%d", userID, time.Now().Unix()),
			"phone":    "",
			"avatar":   "",
			"open_id":  "",
			"union_id": "",
		}).Error
	})
}

// ========== 缓存管理 ==========

func ClearCache(cacheType string) error {
	ctx := context.Background()
	switch cacheType {
	case "all":
		return global.Cache.FlushDB(ctx)
	default:
		keys, err := global.Cache.Keys(ctx, cacheType+"*")
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			return global.Cache.Del(ctx, keys...)
		}
		return nil
	}
}

func GetCacheStats() map[string]interface{} {
	info, _ := global.Cache.Info(context.Background())
	dbSize, _ := global.Cache.DBSize(context.Background())
	return map[string]interface{}{"db_size": dbSize, "info": info}
}
