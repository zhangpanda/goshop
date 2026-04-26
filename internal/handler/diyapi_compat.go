package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/auth"
	"github.com/zhangpanda/goshop/pkg/response"
)

// diyApiAuth accepts admin token from Authorization header OR ?token= query param
func diyApiAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			response.Fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(token, global.Cfg.JWT.Secret)
		if err != nil || !claims.IsAdmin {
			response.Fail(c, http.StatusUnauthorized, "token无效")
			c.Abort()
			return
		}
		c.Set("admin_id", claims.UserID)
		c.Next()
	}
}

// SetupDiyApiCompat 注册 diyapi 和 attachmentapi 兼容路由（供 shopxo-diy/form 前端调用）
func SetupDiyApiCompat(r *gin.Engine) {
	g := r.Group("/api").Use(diyApiAuth())
	{
		// diyapi
		g.POST("/diyapi/init", diyApiInit)
		g.POST("/diyapi/diydetail", diyApiDetail)
		g.POST("/diyapi/diysave", diyApiSave)
		g.POST("/diyapi/diylist", diyApiList)
		g.POST("/diyapi/goodsinit", diyApiGoodsInit)
		g.POST("/diyapi/goodslist", diyApiGoodsList)
		g.POST("/diyapi/goodsautodata", DiyApiGoodsAutoData)
		g.POST("/diyapi/articleautodata", DiyApiArticleAutoData)
		g.POST("/diyapi/articlelist", diyApiArticleList)
		g.POST("/diyapi/brandautodata", DiyApiBrandAutoData)
		g.POST("/diyapi/brandlist", diyApiBrandList)
		g.POST("/diyapi/linkinit", diyApiLinkInit)
		g.POST("/diyapi/customviewlist", diyApiCustomViewList)
		g.POST("/diyapi/designlist", diyApiDesignList)
		g.POST("/diyapi/goodsmagicinit", diyApiGoodsInit)
		g.POST("/diyapi/custominit", diyApiInit)
		g.POST("/diyapi/diyupload", diyApiUpload)
		g.POST("/diyapi/diydownload", diyApiDetail)
		g.POST("/diyapi/articleappointdata", DiyApiArticleAutoData)

		// attachmentapi
		g.POST("/attachmentapi/list", attachmentApiList)
		g.POST("/attachmentapi/upload", attachmentApiUpload)
		g.POST("/attachmentapi/delete", attachmentApiDelete)
		g.POST("/attachmentapi/category", attachmentApiCategory)
		g.POST("/attachmentapi/categorysave", attachmentApiCategorySave)
		g.POST("/attachmentapi/categorydelete", attachmentApiCategoryDelete)
		g.POST("/attachmentapi/save", attachmentApiSave)

		// forminputapi
		g.POST("/forminputapi/init", formInputApiInit)
		g.POST("/forminputapi/forminputdetail", formInputApiDetail)
		g.POST("/forminputapi/forminputsave", formInputApiSave)

		// region/all - 全部地区数据（shopxo-form用）
		g.POST("/region/all", regionAll)
		g.GET("/region/all", regionAll)

		// attachmentapi 补充
		g.POST("/attachmentapi/catch", attachmentApiCatch)
		g.POST("/attachmentapi/movecategory", attachmentApiSave)
		g.POST("/attachmentapi/scanuploaddata", attachmentApiList)

		// forminputapi 补充
		g.POST("/forminputapi/forminputupload", diyApiUpload)
		g.POST("/forminputapi/forminputdownload", formInputApiDetail)
		g.POST("/forminputapi/forminputinstall", formInputApiSave)
		g.POST("/forminputapi/forminputmarket", formInputApiInit)

		// forminputdata
		g.POST("/forminputdata/save", formInputDataSave)

		// diyapi 补充
		g.POST("/diyapi/diymarket", diyApiInit)
		g.POST("/diyapi/diyinstall", diyApiSave)
	}
}

// ========== diyapi ==========

func baseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

func diyApiInit(c *gin.Context) {
	var cats []model.Category
	global.DB.Where("status = 1").Order("sort DESC, id").Find(&cats)
	var attachCats []model.AttachmentCategory
	global.DB.Find(&attachCats)
	var articleCats []model.ArticleCategory
	global.DB.Find(&articleCats)
	var brandCats []model.BrandCategory
	global.DB.Find(&brandCats)
	var brands []model.Brand
	global.DB.Where("status = 1").Find(&brands)

	base := baseURL(c)

	response.OK(c, gin.H{
		"config": map[string]interface{}{
			"attachment_host":  base,
			"public_host":      base + "/",
			"currency_symbol":  "¥",
			"diy_detail_url":   "diyapi/diydetail",
			"diy_save_url":     "diyapi/diysave",
			"diy_upload_url":   "diyapi/diyupload",
			"diy_download_url": "diyapi/diydownload",
			"diy_install_url":  "diyapi/diyinstall",
			"diy_market_url":   "diyapi/diymarket",
			"ueditor": map[string]interface{}{
				"image_suffix": []string{"png", "jpg", "jpeg", "gif", "bmp", "webp"},
				"video_suffix": []string{"mp4", "avi", "wmv", "rm", "rmvb", "mkv"},
				"file_suffix":  []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf", "txt", "zip", "rar"},
			},
			"attachment_category_operate": map[string]int{"is_add": 1, "is_edit": 1, "is_del": 1},
			"attachment_operate":          map[string]int{"is_move": 1, "is_upload": 1, "is_edit": 1, "is_del": 1},
			"diy_config_operate": map[string]int{
				"is_base_data":         1,
				"is_upload_admin":      1,
				"is_save_button":       1,
				"is_save_close_button": 1,
			},
		},
		"attachment_category": attachCats,
		"article_category":    articleCats,
		"brand_category":      brandCats,
		"brand_list":          brands,
		"goods_category":      cats,
		"page_link_list":      []interface{}{},
		"module_list": []map[string]interface{}{
			{"name": "基础组件", "key": "base", "data": []map[string]string{
				{"key": "tabs", "name": "选项卡"},
				{"key": "carousel", "name": "轮播"},
				{"key": "search", "name": "搜索"},
				{"key": "notice", "name": "公告"},
				{"key": "nav-group", "name": "导航组"},
				{"key": "goods-list", "name": "商品列表"},
				{"key": "goods-tabs", "name": "商品标签"},
				{"key": "article-list", "name": "文章列表"},
				{"key": "img-magic", "name": "魔方图片"},
				{"key": "hot-zone", "name": "热区"},
				{"key": "video", "name": "视频"},
				{"key": "custom", "name": "自定义"},
			}},
			{"name": "插件", "key": "plugins", "data": []interface{}{}},
			{"name": "工具组件", "key": "tool", "data": []map[string]string{
				{"key": "title", "name": "标题"},
				{"key": "auxiliary-blank", "name": "辅助空白"},
				{"key": "row-line", "name": "分割线"},
				{"key": "rich-text", "name": "富文本"},
				{"key": "float-window", "name": "浮窗"},
			}},
		},
		"goods_order_by_type_list":   []map[string]string{{"name": "综合", "value": "default"}, {"name": "销量", "value": "sales"}, {"name": "价格", "value": "price"}},
		"article_order_by_type_list": []map[string]string{{"name": "综合", "value": "default"}, {"name": "最新", "value": "new"}},
		"brand_order_by_type_list":   []map[string]string{{"name": "综合", "value": "default"}, {"name": "最新", "value": "new"}},
		"data_order_by_rule_list":    []map[string]string{{"name": "降序", "value": "desc"}, {"name": "升序", "value": "asc"}},
		"plugins":                    []interface{}{},
	})
}

func diyApiDetail(c *gin.Context) {
	id := getCompactID(c)
	var diy model.Diy
	if err := global.DB.First(&diy, id).Error; err != nil {
		response.OK(c, gin.H{"data": nil})
		return
	}
	var config interface{}
	json.Unmarshal([]byte(diy.Data), &config)
	response.OK(c, gin.H{
		"data": gin.H{
			"id":           diy.ID,
			"name":         diy.Name,
			"logo":         "",
			"describe":     "",
			"config":       config,
			"is_enable":    diy.Status,
			"access_count": diy.AccessCount,
			"add_time":     diy.CreatedAt.Format("2006-01-02 15:04:05"),
			"upd_time":     diy.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func diyApiSave(c *gin.Context) {
	var raw map[string]interface{}
	c.ShouldBindJSON(&raw)
	id := parseIDField(raw["id"])
	name, _ := raw["name"].(string)
	isEnable, hasEnable := raw["is_enable"]
	// 兼容: 前端可能发 config 或 data
	configData := raw["config"]
	if configData == nil {
		configData = raw["data"]
	}
	configBytes, _ := json.Marshal(configData)
	if name == "" {
		name = "DIY装修" + time.Now().Format("20060102")
	}
	if id > 0 {
		updates := map[string]interface{}{"name": name, "data": string(configBytes)}
		if hasEnable {
			updates["status"] = isEnable
		}
		global.DB.Model(&model.Diy{}).Where("id = ?", id).Updates(updates)
	} else {
		diy := model.Diy{Name: name, Data: string(configBytes)}
		global.DB.Create(&diy)
		id = diy.ID
	}
	// ShopXO returns data as int (the record ID)
	response.OK(c, id)
}

func diyApiList(c *gin.Context) {
	var list []model.Diy
	global.DB.Order("id DESC").Find(&list)
	response.OK(c, list)
}

func diyApiGoodsInit(c *gin.Context) {
	var cats []model.Category
	global.DB.Where("status = 1").Order("sort DESC, id").Find(&cats)
	response.OK(c, gin.H{"goods_category": cats})
}

func diyApiGoodsList(c *gin.Context) {
	var req struct {
		Keywords   string `json:"keywords"`
		CategoryID uint   `json:"category_id"`
	}
	c.ShouldBindJSON(&req)
	db := global.DB.Model(&model.Goods{}).Where("status = 1")
	if req.Keywords != "" {
		db = db.Where("title LIKE ?", "%"+req.Keywords+"%")
	}
	if req.CategoryID > 0 {
		db = db.Where("category_id = ?", req.CategoryID)
	}
	var list []model.Goods
	db.Preload("SKUs").Order("id DESC").Limit(50).Find(&list)
	response.OK(c, list)
}

func diyApiArticleList(c *gin.Context) {
	var list []model.Article
	global.DB.Where("status = 1").Order("id DESC").Limit(50).Find(&list)
	response.OK(c, list)
}

func diyApiBrandList(c *gin.Context) {
	var list []model.Brand
	global.DB.Where("status = 1").Order("sort DESC, id").Find(&list)
	response.OK(c, list)
}

func diyApiLinkInit(c *gin.Context) {
	response.OK(c, gin.H{
		"link_type": []map[string]string{
			{"name": "商品详情", "value": "goods_detail"},
			{"name": "商品分类", "value": "goods_category"},
			{"name": "自定义页面", "value": "custom_view"},
			{"name": "外部链接", "value": "url"},
		},
	})
}

func diyApiCustomViewList(c *gin.Context) {
	var list []model.CustomView
	global.DB.Where("status = 1").Find(&list)
	response.OK(c, list)
}

func diyApiDesignList(c *gin.Context) {
	var list []model.Design
	global.DB.Find(&list)
	response.OK(c, list)
}

func diyApiUpload(c *gin.Context) {
	Upload(c)
}

// ========== attachmentapi ==========

func attachmentApiList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "24"))
	catID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 64)
	// Also check POST body
	if page == 1 {
		if v, _ := strconv.Atoi(c.PostForm("page")); v > 0 {
			page = v
		}
	}
	if catID == 0 {
		if v, _ := strconv.ParseUint(c.PostForm("category_id"), 10, 64); v > 0 {
			catID = v
		}
	}
	db := global.DB.Model(&model.Attachment{})
	if catID > 0 {
		db = db.Where("category_id = ?", catID)
	}
	var total int64
	db.Count(&total)
	var list []model.Attachment
	db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "data": list, "page": page, "page_size": pageSize})
}

func attachmentApiUpload(c *gin.Context) { Upload(c) }

func attachmentApiDelete(c *gin.Context) {
	var req struct {
		ID  uint   `json:"id"`
		IDs string `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if req.ID > 0 {
		global.DB.Delete(&model.Attachment{}, req.ID)
	} else if req.IDs != "" {
		global.DB.Where("id IN (?)", req.IDs).Delete(&model.Attachment{})
	}
	response.OK(c, nil)
}

func attachmentApiCategory(c *gin.Context) {
	var list []model.AttachmentCategory
	global.DB.Find(&list)
	response.OK(c, list)
}

func attachmentApiCategorySave(c *gin.Context) {
	var req model.AttachmentCategory
	c.ShouldBindJSON(&req)
	if req.ID > 0 {
		global.DB.Model(&req).Updates(req)
	} else {
		global.DB.Create(&req)
	}
	response.OK(c, req)
}

func attachmentApiCategoryDelete(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	c.ShouldBindJSON(&req)
	global.DB.Delete(&model.AttachmentCategory{}, req.ID)
	response.OK(c, nil)
}

func attachmentApiSave(c *gin.Context) {
	var req struct {
		ID         uint `json:"id"`
		CategoryID uint `json:"category_id"`
	}
	c.ShouldBindJSON(&req)
	global.DB.Model(&model.Attachment{}).Where("id = ?", req.ID).Update("category_id", req.CategoryID)
	response.OK(c, nil)
}

func attachmentApiCatch(c *gin.Context) {
	var req struct {
		Type       string      `json:"type"`
		Source     interface{} `json:"source"`
		CategoryID uint        `json:"category_id"`
	}
	c.ShouldBindJSON(&req)
	// Remote image catch is complex; return stub for now
	response.OK(c, []interface{}{})
}

// ========== forminputapi ==========

func formInputApiInit(c *gin.Context) {
	base := baseURL(c)
	var attachCats []model.AttachmentCategory
	global.DB.Find(&attachCats)

	response.OK(c, gin.H{
		"config": map[string]interface{}{
			"attachment_host":        base,
			"public_host":            base + "/",
			"currency_symbol":        "¥",
			"forminput_detail_url":   "forminputapi/forminputdetail",
			"forminput_save_url":     "forminputapi/forminputsave",
			"forminput_upload_url":   "forminputapi/forminputupload",
			"forminput_download_url": "forminputapi/forminputdownload",
			"forminput_install_url":  "forminputapi/forminputinstall",
			"forminput_market_url":   "forminputapi/forminputmarket",
			"ueditor": map[string]interface{}{
				"image_suffix": []string{"png", "jpg", "jpeg", "gif", "bmp", "webp"},
				"video_suffix": []string{"mp4", "avi", "wmv"},
				"file_suffix":  []string{"doc", "docx", "xls", "xlsx", "pdf", "txt", "zip", "rar"},
			},
			"attachment_category_operate": map[string]int{"is_add": 1, "is_edit": 1, "is_del": 1},
			"attachment_operate":          map[string]int{"is_move": 1, "is_upload": 1, "is_edit": 1, "is_del": 1},
			"forminput_config_operate": map[string]int{
				"is_base_data": 1, "is_mode_switch": 1, "is_common_config": 1,
				"is_forminput_config": 1, "is_submit_button": 1,
				"is_save_button": 1, "is_save_close_button": 1,
			},
		},
		"attachment_category": attachCats,
		"module_list": []map[string]interface{}{
			{"name": "基础组件", "key": "base", "data": []map[string]string{
				{"key": "single-text", "name": "单行文本"},
				{"key": "multi-text", "name": "多行文本"},
				{"key": "number", "name": "数字"},
				{"key": "radio-btns", "name": "单选"},
				{"key": "checkbox", "name": "多选"},
				{"key": "select", "name": "下拉单选"},
				{"key": "date", "name": "日期"},
			}},
			{"name": "高级组件", "key": "senior", "data": []map[string]string{
				{"key": "upload-img", "name": "上传图片"},
				{"key": "upload-video", "name": "上传视频"},
				{"key": "upload-attachments", "name": "上传附件"},
				{"key": "rich-text", "name": "富文本"},
				{"key": "address", "name": "地址"},
				{"key": "phone", "name": "手机号"},
				{"key": "score", "name": "评分"},
			}},
			{"name": "扩展组件", "key": "extends", "data": []map[string]string{
				{"key": "text", "name": "文字说明"},
				{"key": "img", "name": "图片"},
				{"key": "auxiliary-line", "name": "辅助线"},
			}},
		},
	})
}

func formInputApiDetail(c *gin.Context) {
	id := getCompactID(c)
	var form model.FormInput
	if err := global.DB.First(&form, id).Error; err != nil {
		response.OK(c, gin.H{"data": nil})
		return
	}
	var config interface{}
	json.Unmarshal([]byte(form.Config), &config)
	response.OK(c, gin.H{
		"data": gin.H{
			"id":        form.ID,
			"name":      form.Name,
			"logo":      "",
			"describe":  "",
			"config":    config,
			"is_enable": form.Status,
			"add_time":  form.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func formInputApiSave(c *gin.Context) {
	var raw map[string]interface{}
	c.ShouldBindJSON(&raw)
	id := parseIDField(raw["id"])
	name, _ := raw["name"].(string)
	isEnable, hasEnable := raw["is_enable"]
	configBytes, _ := json.Marshal(raw["config"])
	if name == "" {
		name = "Form表单" + time.Now().Format("20060102")
	}
	if id > 0 {
		updates := map[string]interface{}{"name": name, "config": string(configBytes)}
		if hasEnable {
			updates["status"] = isEnable
		}
		global.DB.Model(&model.FormInput{}).Where("id = ?", id).Updates(updates)
	} else {
		form := model.FormInput{Name: name, Config: string(configBytes)}
		global.DB.Create(&form)
		id = form.ID
	}
	response.OK(c, id)
}

// parseIDField converts interface{} (string or number) to uint
func parseIDField(v interface{}) uint {
	switch id := v.(type) {
	case float64:
		return uint(id)
	case string:
		if n, err := strconv.ParseUint(id, 10, 64); err == nil {
			return uint(n)
		}
	}
	return 0
}

// getCompactID extracts ID from JSON body or query params
func getCompactID(c *gin.Context) uint {
	// Try query param first (doesn't consume body)
	if v, _ := strconv.ParseUint(c.Query("id"), 10, 64); v > 0 {
		return uint(v)
	}
	// Try JSON body - accept both number and string
	var body map[string]interface{}
	if c.ShouldBindJSON(&body) == nil {
		if v, ok := body["id"]; ok {
			switch id := v.(type) {
			case float64:
				return uint(id)
			case string:
				if n, err := strconv.ParseUint(id, 10, 64); err == nil {
					return uint(n)
				}
			}
		}
	}
	return 0
}

// regionAll 返回全部地区数据（三级嵌套，优化为3次查询）
func regionAll(c *gin.Context) {
	var all []model.Region
	global.DB.Order("sort, id").Find(&all)

	type RegionNode struct {
		ID       uint         `json:"id"`
		Name     string       `json:"name"`
		Level    int8         `json:"level"`
		Children []RegionNode `json:"children,omitempty"`
	}

	// Build lookup by parent_id
	byParent := map[uint][]model.Region{}
	for _, r := range all {
		byParent[r.ParentID] = append(byParent[r.ParentID], r)
	}

	var result []RegionNode
	for _, p := range byParent[0] {
		pn := RegionNode{ID: p.ID, Name: p.Name, Level: p.Level}
		for _, city := range byParent[p.ID] {
			cn := RegionNode{ID: city.ID, Name: city.Name, Level: city.Level}
			for _, d := range byParent[city.ID] {
				cn.Children = append(cn.Children, RegionNode{ID: d.ID, Name: d.Name, Level: d.Level})
			}
			pn.Children = append(pn.Children, cn)
		}
		result = append(result, pn)
	}
	response.OK(c, result)
}

// formInputDataSave 表单数据提交
func formInputDataSave(c *gin.Context) {
	var req struct {
		FormID uint        `json:"form_id"`
		Data   interface{} `json:"data"`
	}
	c.ShouldBindJSON(&req)
	dataBytes, _ := json.Marshal(req.Data)
	record := model.FormInputData{FormID: req.FormID, Data: string(dataBytes)}
	global.DB.Create(&record)
	response.OK(c, gin.H{"id": record.ID})
}

// Ensure fmt is used
var _ = fmt.Sprintf
