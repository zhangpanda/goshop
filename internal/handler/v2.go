package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 统计 ==========

func AdminStatistical(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	response.OK(c, service.GetStatistical(days))
}

// ========== DiyApi ==========

func DiyApiGoodsAutoData(c *gin.Context) {
	var p service.DiyApiParams
	if !BindQuery(c, &p) {
		return
	}
	list, err := service.DiyApiGoodsAutoData(&p)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func DiyApiArticleAutoData(c *gin.Context) {
	catID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	list, _ := service.DiyApiArticleAutoData(uint(catID), limit)
	response.OK(c, list)
}

func DiyApiBrandAutoData(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, _ := service.DiyApiBrandAutoData(limit)
	response.OK(c, list)
}

func DiyApiGoodsFavorAutoData(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	list, _ := service.DiyApiGoodsFavorAutoData(c.GetUint("user_id"), limit)
	response.OK(c, list)
}

func DiyApiGoodsBrowseAutoData(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	list, _ := service.DiyApiGoodsBrowseAutoData(c.GetUint("user_id"), limit)
	response.OK(c, list)
}

// ========== 小程序管理 ==========

func SaveAppMini(c *gin.Context) {
	var req service.AppMiniReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.SaveAppMini(&req); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

func GetAppMiniList(c *gin.Context) {
	list, _ := service.GetAppMiniList()
	response.OK(c, list)
}

func DeleteAppMini(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	service.DeleteAppMini(uint(id))
	response.OK(c, nil)
}

// ========== 站点配置 ==========

func GetSiteConfigHandler(c *gin.Context) {
	response.OK(c, service.GetSiteConfig())
}

func SaveSiteConfigHandler(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.SaveSiteConfig(req)
	response.OK(c, nil)
}

func GetSelfExtractionAddress(c *gin.Context) {
	response.OK(c, service.GetSelfExtractionAddressList())
}

func SaveSelfExtractionAddress(c *gin.Context) {
	var req []map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.SaveSelfExtractionAddress(req)
	response.OK(c, nil)
}

// ========== 动态表格 ==========

func FormTableQueryHandler(c *gin.Context) {
	var req service.FormTableParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	total, data, err := service.FormTableQuery(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": data})
}

// ========== 二维码 ==========

func GenerateQRCode(c *gin.Context) {
	content := c.Query("content")
	if content == "" {
		response.Fail(c, http.StatusBadRequest, "content不能为空")
		return
	}
	response.OK(c, gin.H{"url": service.GenerateQRCodeURL(content)})
}

// ========== SQL控制台 ==========

func SqlConsoleExecute(c *gin.Context) {
	// 配置开关：默认关闭
	if !app.Must().Cfg.Server.SqlConsole {
		response.Fail(c, http.StatusForbidden, "SQL控制台未启用（需配置 server.sql_console: true）")
		return
	}
	// 仅超级管理员可用
	if !service.IsSuperAdmin(c.GetUint("admin_role_id")) {
		response.Fail(c, http.StatusForbidden, "仅超级管理员可使用SQL控制台")
		return
	}
	var req struct {
		SQL string `json:"sql" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	results, err := service.ExecuteSQL(req.SQL)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, results)
}

// ========== 系统信息 ==========

func GetSystemInfo(c *gin.Context) {
	response.OK(c, service.GetSystemInfo())
}
