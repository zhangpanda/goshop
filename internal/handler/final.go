package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// Design
func DesignListHandler(c *gin.Context) { l, _ := service.DesignList(); response.OK(c, l) }
func DesignCreateHandler(c *gin.Context) {
	var r struct {
		Name string `json:"name" binding:"required"`
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	d, _ := service.DesignCreate(r.Name, r.Data)
	response.OK(c, d)
}
func DesignUpdateHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var r struct {
		Data string `json:"data"`
	}
	if !BindJSON(c, &r) {
		return
	}
	service.DesignUpdate(uint(id), r.Data)
	response.OK(c, nil)
}

// Layout
func LayoutListHandler(c *gin.Context) { l, _ := service.LayoutList(); response.OK(c, l) }
func LayoutSaveHandler(c *gin.Context) {
	var r struct {
		Name string `json:"name"`
		Type string `json:"type" binding:"required"`
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.LayoutSave(r.Name, r.Type, r.Data)
	response.OK(c, nil)
}

// GoodsContentApp
func SaveGoodsContentAppHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var r struct {
		Content string `json:"content"`
	}
	if !BindJSON(c, &r) {
		return
	}
	service.SaveGoodsContentApp(uint(id), r.Content)
	response.OK(c, nil)
}
func GetGoodsContentAppHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	response.OK(c, gin.H{"content": service.GetGoodsContentApp(uint(id))})
}

// OrderService
func CreateOrderServiceHandler(c *gin.Context) {
	var r struct {
		OrderID uint   `json:"order_id" binding:"required"`
		Type    string `json:"type"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s, _ := service.CreateOrderService(r.OrderID, c.GetUint("user_id"), r.Type, r.Content)
	response.OK(c, s)
}
func GetOrderServiceList(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	l, _ := service.OrderServiceList(uint(id))
	response.OK(c, l)
}
func AdminReplyOrderService(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var r struct {
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.ReplyOrderService(uint(id), c.GetUint("admin_id"), r.Reply)
	response.OK(c, nil)
}
func AdminOrderServiceList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		sv := int8(v)
		status = &sv
	}
	l, total, _ := service.AdminOrderServiceList(status, page, 20)
	response.OK(c, gin.H{"total": total, "list": l})
}

// QuickNav
func QuickNavListHandler(c *gin.Context) { l, _ := service.QuickNavList(); response.OK(c, l) }
func QuickNavCreateHandler(c *gin.Context) {
	var n model.QuickNav
	if !BindJSON(c, &n) {
		return
	}
	service.QuickNavCreate(&n)
	response.OK(c, n)
}

// PluginsDataConfig
func PluginConfigGetHandler(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Query("plugin_id"), 10, 64)
	response.OK(c, gin.H{"value": service.PluginConfigGet(uint(pid), c.Query("key"))})
}
func PluginConfigSetHandler(c *gin.Context) {
	var r struct {
		PluginID uint   `json:"plugin_id"`
		Key      string `json:"key"`
		Value    string `json:"value"`
	}
	if !BindJSON(c, &r) {
		return
	}
	service.PluginConfigSet(r.PluginID, r.Key, r.Value)
	response.OK(c, nil)
}

// RolePlugins
func GetRolePluginsHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	ids, err := service.GetRolePluginIDs(uint(id))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, ids)
}

func SaveRolePluginsHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var r struct {
		PluginIDs []uint `json:"plugin_ids"`
	}
	if !BindJSON(c, &r) {
		return
	}
	service.SaveRolePlugins(uint(id), r.PluginIDs)
	response.OK(c, nil)
}

// FormFields
func SaveFormFieldsHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var r struct {
		Fields []model.FormTableUserFields `json:"fields"`
	}
	if !BindJSON(c, &r) {
		return
	}
	service.SaveFormFields(uint(id), r.Fields)
	response.OK(c, nil)
}
func GetFormFieldsHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	l, _ := service.GetFormFields(uint(id))
	response.OK(c, l)
}
