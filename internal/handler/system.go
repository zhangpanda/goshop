package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ---- 品牌 ----
func CreateBrand(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Logo string `json:"logo"`
		Desc string `json:"desc"`
		Sort int    `json:"sort"`
	}
	if !BindJSON(c, &req) {
		return
	}
	b, err := service.CreateBrand(req.Name, req.Logo, req.Desc, req.Sort)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, b)
}
func GetBrandList(c *gin.Context) {
	list, err := service.GetBrandList()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

// ---- 文章 ----
func CreateArticleCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}
	if !BindJSON(c, &req) {
		return
	}
	cat, err := service.CreateArticleCategory(req.Name, req.Sort)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, cat)
}
func GetArticleCategoryList(c *gin.Context) {
	list, err := service.GetArticleCategoryList()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}
func CreateArticle(c *gin.Context) {
	var req service.ArticleReq
	if !BindJSON(c, &req) {
		return
	}
	a, err := service.CreateArticle(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, a)
}
func GetArticleList(c *gin.Context) {
	catID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	page, pageSize := QueryPage(c)
	list, total, err := service.GetArticleList(uint(catID), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}
func GetArticleDetail(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	a, err := service.GetArticleDetail(uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}
	response.OK(c, a)
}

// ---- 规格模板 ----
func CreateSpecTemplate(c *gin.Context) {
	var req service.SpecTemplateReq
	if !BindJSON(c, &req) {
		return
	}
	t, err := service.CreateSpecTemplate(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, t)
}
func GetSpecTemplateList(c *gin.Context) {
	list, err := service.GetSpecTemplateList()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

// ---- 参数模板 ----
func CreateParamsTemplate(c *gin.Context) {
	var req service.ParamsTemplateReq
	if !BindJSON(c, &req) {
		return
	}
	t, err := service.CreateParamsTemplate(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, t)
}
func GetParamsTemplateList(c *gin.Context) {
	list, err := service.GetParamsTemplateList()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}
func SaveGoodsParams(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Params []service.ParamsConfigItem `json:"params"`
	}
	if !BindJSON(c, &req) {
		return
	}
	service.SaveGoodsParams(uint(id), req.Params)
	response.OK(c, nil)
}
func GetGoodsParamsHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	list, _ := service.GetGoodsParams(uint(id))
	response.OK(c, list)
}

// ---- 搜索 ----
func GetHotKeywords(c *gin.Context) {
	kw, _ := service.GetHotKeywords(10)
	response.OK(c, kw)
}
func GetSearchHistory(c *gin.Context) {
	list, _ := service.GetSearchHistory(c.GetUint("user_id"))
	response.OK(c, list)
}
func ClearSearchHistory(c *gin.Context) {
	service.ClearSearchHistory(c.GetUint("user_id"))
	response.OK(c, nil)
}
func GetScreeningPrices(c *gin.Context) {
	list, _ := service.GetScreeningPrices()
	response.OK(c, list)
}

// ---- 系统配置 ----
func GetConfigGroup(c *gin.Context) {
	list, _ := service.GetConfigGroup(c.Query("group"))
	type item struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Desc  string `json:"desc"`
	}
	items := make([]item, len(list))
	for i, cfg := range list {
		items[i] = item{Key: cfg.Key, Value: cfg.Value, Desc: cfg.Desc}
	}
	response.OK(c, items)
}
func SetConfigHandler(c *gin.Context) {
	var req struct {
		Group   string            `json:"group"`
		Configs map[string]string `json:"configs"`
		// 兼容单条
		Key   string `json:"key"`
		Value string `json:"value"`
		Desc  string `json:"desc"`
	}
	if !BindJSON(c, &req) {
		return
	}
	if len(req.Configs) > 0 {
		for k, v := range req.Configs {
			service.SetConfig(k, v, req.Group, "")
		}
	} else if req.Key != "" {
		service.SetConfig(req.Key, req.Value, req.Group, req.Desc)
	}
	response.OK(c, nil)
}

// ---- 地区 ----
func GetRegionList(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.DefaultQuery("parent_id", "0"), 10, 64)
	list, _ := service.GetRegionList(uint(pid))
	response.OK(c, list)
}

// ---- 幻灯片/导航/链接 ----
func GetSlideList(c *gin.Context) { list, _ := service.SlideList(); response.OK(c, list) }
func CreateSlideHandler(c *gin.Context) {
	var s model.Slide
	if !BindJSON(c, &s) {
		return
	}
	service.CreateSlide(&s)
	response.OK(c, s)
}
func GetNavigationList(c *gin.Context) {
	list, _ := service.NavigationList(c.Query("type"))
	response.OK(c, list)
}
func CreateNavigationHandler(c *gin.Context) {
	var n model.Navigation
	if !BindJSON(c, &n) {
		return
	}
	service.CreateNavigation(&n)
	response.OK(c, n)
}
func GetLinkList(c *gin.Context) { list, _ := service.LinkList(); response.OK(c, list) }
func CreateLinkHandler(c *gin.Context) {
	var l model.Link
	if !BindJSON(c, &l) {
		return
	}
	service.CreateLink(&l)
	response.OK(c, l)
}

// ---- 支付方式 ----
func GetPaymentList(c *gin.Context) { list, _ := service.PaymentList(); response.OK(c, list) }
func CreatePaymentHandler(c *gin.Context) {
	var p model.Payment
	if !BindJSON(c, &p) {
		return
	}
	service.CreatePayment(&p)
	response.OK(c, p)
}

// ---- 附件 ----
func GetAttachmentList(c *gin.Context) {
	catID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	page, pageSize := QueryPage(c)
	list, total, _ := service.AttachmentList(uint(catID), page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}
func GetAttachmentCategoryList(c *gin.Context) {
	list, _ := service.AttachmentCategoryList()
	response.OK(c, list)
}
func CreateAttachmentCategoryHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if !BindJSON(c, &req) {
		return
	}
	service.CreateAttachmentCategory(req.Name)
	response.OK(c, nil)
}

// ---- 错误日志 ----
func GetErrorLogList(c *gin.Context) {
	page, pageSize := QueryPage(c)
	list, total, _ := service.GetErrorLogList(page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

// ---- 订单状态历史 ----
func GetOrderStatusHistory(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	list, _ := service.GetOrderStatusHistory(uint(id))
	response.OK(c, list)
}
