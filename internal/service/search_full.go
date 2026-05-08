package service

import (
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
)

// SearchStartData 搜索页初始化数据（筛选面板需要的所有选项）
type SearchStartResp struct {
	Brands     []model.Brand          `json:"brands"`
	Categories []model.Category       `json:"categories"`
	Prices     []model.ScreeningPrice `json:"prices"`
	SpecValues []SpecFilterItem       `json:"spec_values"`
	Params     []ParamFilterItem      `json:"params"`
	Regions    []string               `json:"regions"`
	HotWords   []string               `json:"hot_words"`
	MaxPrice   int64                  `json:"max_price"`
}
type SpecFilterItem struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}
type ParamFilterItem struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func SearchStartData() *SearchStartResp {
	r := &SearchStartResp{}
	app.Must().DB.Where("status = 1").Order("sort DESC").Find(&r.Brands)
	r.Categories = GoodsCategoryAll()
	app.Must().DB.Order("sort").Find(&r.Prices)
	r.HotWords, _ = GetHotKeywords(10)
	r.SpecValues = SearchGoodsSpecValueList()
	r.Params = SearchGoodsParamsValueList()
	r.Regions = SearchGoodsProduceRegionList()
	app.Must().DB.Model(&model.GoodsSKU{}).Select("COALESCE(MAX(price),0)").Scan(&r.MaxPrice)
	return r
}

// SearchGoodsSpecValueList 获取所有可筛选的规格值
func SearchGoodsSpecValueList() []SpecFilterItem {
	var types []model.SpecType
	app.Must().DB.Preload("Values").Find(&types)
	var result []SpecFilterItem
	for _, t := range types {
		vals := make([]string, len(t.Values))
		for i, v := range t.Values {
			vals[i] = v.Value
		}
		if len(vals) > 0 {
			result = append(result, SpecFilterItem{Name: t.Name, Values: vals})
		}
	}
	return result
}

// SearchGoodsParamsValueList 获取所有可筛选的参数值
func SearchGoodsParamsValueList() []ParamFilterItem {
	type row struct{ Name, Value string }
	var rows []row
	app.Must().DB.Model(&model.GoodsParams{}).Select("DISTINCT name, value").Find(&rows)
	m := map[string][]string{}
	for _, r := range rows {
		m[r.Name] = append(m[r.Name], r.Value)
	}
	var result []ParamFilterItem
	for k, v := range m {
		result = append(result, ParamFilterItem{Name: k, Values: v})
	}
	return result
}

// SearchGoodsProduceRegionList 获取所有产地列表
func SearchGoodsProduceRegionList() []string {
	var vals []string
	app.Must().DB.Model(&model.GoodsParams{}).Where("name = '产地'").Distinct("value").Pluck("value", &vals)
	return vals
}

// CategoryBrandList 指定分类下的品牌列表
func CategoryBrandList(categoryID uint) []model.Brand {
	ids := GoodsCategoryItemsIds([]uint{categoryID}, 3)
	var brandIDs []uint
	app.Must().DB.Model(&model.Goods{}).Where("category_id IN ? AND status = 1", ids).Distinct("brand_id").Where("brand_id > 0").Pluck("brand_id", &brandIDs)
	if len(brandIDs) == 0 {
		return nil
	}
	var list []model.Brand
	app.Must().DB.Where("id IN ? AND status = 1", brandIDs).Find(&list)
	return list
}

// SearchProhibitCheck 搜索禁止词检查
func SearchProhibitCheck(keyword string) bool {
	raw := GetConfig("search_prohibit_keywords")
	if raw == "" {
		return false
	}
	for _, w := range splitLines(raw) {
		if w != "" && keyword == w {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

type SearchRankItem struct {
	Keyword string `json:"keyword"`
	Count   int64  `json:"count"`
}

func SearchRankingList(limit int) []SearchRankItem {
	if limit <= 0 {
		limit = 20
	}
	var list []SearchRankItem
	app.Must().DB.Model(&model.SearchHistory{}).Select("keyword, COUNT(*) as count").
		Group("keyword").Order("count DESC").Limit(limit).Find(&list)
	return list
}
