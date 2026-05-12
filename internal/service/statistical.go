package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

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

	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		var ds DayStat
		ds.Date = d
		app.Must().DB.Model(&model.Order{}).Where("DATE(created_at) = ? AND status > 0", d).
			Select("COALESCE(SUM(pay_amount),0) as sales, COUNT(*) as count").Scan(&ds)
		ds.Date = d
		data.Trend = append(data.Trend, ds)
	}

	app.Must().DB.Model(&model.Goods{}).Select("id as goods_id, title, sales_count as sales").
		Order("sales_count DESC").Limit(10).Find(&data.GoodsTop)

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

	app.Must().DB.Model(&model.Order{}).Select("status, COUNT(*) as count").Group("status").Find(&data.OrderDist)

	app.Must().DB.Model(&model.Order{}).Where("status = 0").Count(&data.OrderPendingCount)
	app.Must().DB.Model(&model.OrderAftersale{}).Where("status = 0").Count(&data.AftersalePendingCount)
	app.Must().DB.Model(&model.Goods{}).Where("status = 0").Count(&data.GoodsOfflineCount)
	app.Must().DB.Model(&model.Review{}).Where("reply = '' OR reply IS NULL").Count(&data.ReviewPendingCount)

	app.Must().DB.Model(&model.PayLog{}).Select("client_type, COUNT(*) as count, COALESCE(SUM(total_price),0) as amount").
		Where("status = 1").Group("client_type").Find(&data.PayTypeStats)

	app.Must().DB.Raw(`SELECT JSON_UNQUOTE(JSON_EXTRACT(address, '$.province')) as province, COUNT(*) as count 
		FROM orders WHERE status > 0 AND address != '' AND address IS NOT NULL 
		GROUP BY province HAVING province IS NOT NULL AND province != '' 
		ORDER BY count DESC LIMIT 10`).Find(&data.RegionStats)

	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		var ds DayStat
		ds.Date = d
		app.Must().DB.Model(&model.User{}).Where("DATE(created_at) = ?", d).Count(&ds.Count)
		data.NewUserTrend = append(data.NewUserTrend, ds)
	}

	return data
}
