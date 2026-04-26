package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// 插件
func PluginList(c *gin.Context) { l, _ := service.PluginList(); response.OK(c, l) }
func PluginInstall(c *gin.Context) {
	var r struct{ Name, Title, Desc, Author, Version string }
	c.ShouldBindJSON(&r)
	service.PluginInstall(r.Name, r.Title, r.Desc, r.Author, r.Version)
	response.OK(c, nil)
}
func PluginUninstall(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	service.PluginUninstall(uint(id))
	response.OK(c, nil)
}

// DIY
func DiyListHandler(c *gin.Context) { l, _ := service.DiyList(); response.OK(c, l) }
func DiyCreateHandler(c *gin.Context) {
	var r struct {
		Name string `json:"name" binding:"required"`
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	d, _ := service.DiyCreate(r.Name, r.Data)
	response.OK(c, d)
}
func DiyUpdateHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var r struct {
		Data string `json:"data"`
	}
	c.ShouldBindJSON(&r)
	service.DiyUpdate(uint(id), r.Data)
	response.OK(c, nil)
}
func DiyDeleteHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	global.DB.Delete(&model.Diy{}, id)
	response.OK(c, nil)
}

// 自定义页面
func CustomViewListHandler(c *gin.Context) { l, _ := service.CustomViewList(); response.OK(c, l) }
func CustomViewCreateHandler(c *gin.Context) {
	var r struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	v, _ := service.CustomViewCreate(r.Title, r.Content)
	response.OK(c, v)
}

// 主题
func ThemeListHandler(c *gin.Context) { l, _ := service.ThemeList(); response.OK(c, l) }
func ThemeCreateHandler(c *gin.Context) {
	var r struct {
		Name string `json:"name" binding:"required"`
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.ThemeCreate(r.Name, r.Data)
	response.OK(c, nil)
}

// 表单
func FormInputListHandler(c *gin.Context) { l, _ := service.FormInputList(); response.OK(c, l) }
func FormInputCreateHandler(c *gin.Context) {
	var r struct {
		Name   string `json:"name" binding:"required"`
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	form := model.FormInput{Name: r.Name, Config: r.Config, Status: 1}
	global.DB.Create(&form)
	response.OK(c, gin.H{"id": form.ID})
}
func FormInputDeleteHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	global.DB.Delete(&model.FormInput{}, id)
	response.OK(c, nil)
}
func FormInputDataSubmitHandler(c *gin.Context) {
	var r struct {
		FormID uint   `json:"form_id" binding:"required"`
		Data   string `json:"data"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.FormInputDataSubmit(r.FormID, c.GetUint("user_id"), r.Data)
	response.OK(c, nil)
}
func FormInputDataListHandler(c *gin.Context) {
	fid, _ := strconv.ParseUint(c.Query("form_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	l, total, _ := service.FormInputDataList(uint(fid), page, 20)
	response.OK(c, gin.H{"total": total, "list": l})
}

// APP导航
func AppHomeNavListHandler(c *gin.Context) { l, _ := service.AppHomeNavList(); response.OK(c, l) }
func AppHomeNavCreateHandler(c *gin.Context) {
	var n model.AppHomeNav
	c.ShouldBindJSON(&n)
	service.AppHomeNavCreate(&n)
	response.OK(c, n)
}
func AppCenterNavListHandler(c *gin.Context) { l, _ := service.AppCenterNavList(); response.OK(c, l) }
func AppCenterNavCreateHandler(c *gin.Context) {
	var n model.AppCenterNav
	c.ShouldBindJSON(&n)
	service.AppCenterNavCreate(&n)
	response.OK(c, n)
}
func AppTabbarListHandler(c *gin.Context) { l, _ := service.AppTabbarList(); response.OK(c, l) }
func AppTabbarSaveHandler(c *gin.Context) {
	var r struct {
		Items []model.AppTabbar `json:"items"`
	}
	c.ShouldBindJSON(&r)
	service.AppTabbarSave(r.Items)
	response.OK(c, nil)
}

// 快捷菜单
func ShortcutMenuListHandler(c *gin.Context) { l, _ := service.ShortcutMenuList(); response.OK(c, l) }
func ShortcutMenuSaveHandler(c *gin.Context) {
	var r struct {
		Items []model.ShortcutMenu `json:"items"`
	}
	c.ShouldBindJSON(&r)
	service.ShortcutMenuSave(r.Items)
	response.OK(c, nil)
}

// 协议
func AgreementGetHandler(c *gin.Context) { response.OK(c, service.AgreementGet(c.Query("name"))) }
func AgreementSaveHandler(c *gin.Context) {
	var r struct {
		Name    string `json:"name" binding:"required"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.AgreementSave(r.Name, r.Content)
	response.OK(c, nil)
}

// SEO（复用Config）
func SeoGetHandler(c *gin.Context) {
	list, _ := service.GetConfigGroup("seo")
	response.OK(c, list)
}
func SeoSaveHandler(c *gin.Context) {
	var r struct {
		Title    string `json:"title"`
		Keywords string `json:"keywords"`
		Desc     string `json:"desc"`
	}
	c.ShouldBindJSON(&r)
	service.SetConfig("seo_title", r.Title, "seo", "SEO标题")
	service.SetConfig("seo_keywords", r.Keywords, "seo", "SEO关键词")
	service.SetConfig("seo_desc", r.Desc, "seo", "SEO描述")
	response.OK(c, nil)
}
