package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 多语言 ==========

func GetMultilingualConfig(c *gin.Context) {
	response.OK(c, service.GetMultilingualConfig())
}

func SetMultilingualConfig(c *gin.Context) {
	var req struct {
		DefaultLang string   `json:"default_lang" binding:"required"`
		Available   []string `json:"available" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.SetMultilingualConfig(req.DefaultLang, req.Available)
	response.OK(c, nil)
}

func GetLangPack(c *gin.Context) {
	lang := c.DefaultQuery("lang", "zh")
	module := c.DefaultQuery("module", "common")
	data, _ := service.GetLangPack(lang, module)
	response.OK(c, data)
}

// ========== 货币 ==========

func GetCurrencyConfig(c *gin.Context) {
	response.OK(c, service.GetCurrencyConfig())
}

func SetCurrencyConfig(c *gin.Context) {
	var req service.CurrencyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.SetCurrencyConfig(&req)
	response.OK(c, nil)
}

// ========== 订单预约确认 ==========

func AdminBookingConfirm(c *gin.Context) {
	var req struct {
		OrderID uint `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.BookingConfirm(req.OrderID); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

// ========== 数据导出 ==========

func ExportData(c *gin.Context) {
	var req service.ExportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=export_"+req.Type+".csv")
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
	if err := service.ExportData(c.Writer, &req); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
	}
}

// ========== 账号注销 ==========

func UserLogout(c *gin.Context) {
	if err := service.UserLogout(c.GetUint("user_id")); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ========== 缓存管理 ==========

func ClearCache(c *gin.Context) {
	cacheType := c.DefaultQuery("type", "all")
	if err := service.ClearCache(cacheType); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

func GetCacheStats(c *gin.Context) {
	response.OK(c, service.GetCacheStats())
}
