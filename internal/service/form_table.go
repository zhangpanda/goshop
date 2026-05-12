package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zhangpanda/goshop/internal/app"
)

type FormTableParams struct {
	Table         string           `json:"table" binding:"required"`
	Keyword       string           `json:"keyword"`
	KeywordFields string           `json:"keyword_fields"`
	Where         []FormTableWhere `json:"where"`
	OrderBy       string           `json:"order_by"`
	Page          int              `json:"page"`
	PageSize      int              `json:"page_size"`
}
type FormTableWhere struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

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
