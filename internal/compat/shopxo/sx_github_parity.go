package shopxo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/handler"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func sxArticleIndex(c *gin.Context) {
	cats, err := service.GetArticleCategoryList()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]interface{}, 0, len(cats))
	for i := range cats {
		out = append(out, map[string]interface{}{"id": cats[i].ID, "name": cats[i].Name})
	}
	response.OK(c, map[string]interface{}{"category_list": out})
}

func sxArticleDataList(c *gin.Context) {
	catID, _ := strconv.ParseUint(c.PostForm("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultPostForm("page", "1"))
	if page < 1 {
		page = 1
	}
	const pageSize = 20
	list, total, err := service.GetArticleList(uint(catID), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	pt := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pt < 1 {
		pt = 1
	}
	rows := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		a := &list[i]
		rows = append(rows, map[string]interface{}{
			"id": a.ID, "title": a.Title, "cover": a.Cover, "describe": "",
		})
	}
	response.OK(c, map[string]interface{}{
		"data":       rows,
		"total":      total,
		"page_total": pt,
	})
}

func sxArticleDetail(c *gin.Context) {
	id := getID(c)
	a, err := service.GetArticleDetail(id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}
	response.OK(c, map[string]interface{}{
		"data": map[string]interface{}{
			"id": a.ID, "title": a.Title, "content": a.Content, "cover": a.Cover,
			"seo_title": a.Title, "seo_desc": "", "share_images": a.Cover,
		},
		"last_next": nil,
	})
}

func sxPaylogIndex(c *gin.Context) {
	response.OK(c, service.PayLogMenuRowsShopXO(c.GetUint("user_id")))
}

func sxPaylogDetail(c *gin.Context) {
	id := getID(c)
	rows, err := service.PayLogOrderRowsShopXO(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, rows)
}

func sxOrderCommentSave(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if orderID == 0 {
		response.Fail(c, http.StatusBadRequest, "订单无效")
		return
	}
	var goodsIDs []uint
	var ratings []int
	var contents []string
	_ = json.Unmarshal([]byte(c.PostForm("goods_id")), &goodsIDs)
	_ = json.Unmarshal([]byte(c.PostForm("rating")), &ratings)
	_ = json.Unmarshal([]byte(c.PostForm("content")), &contents)
	rawImg := c.PostForm("images")
	var imageJSONs []string
	if rawImg != "" {
		var rawArr []json.RawMessage
		if json.Unmarshal([]byte(rawImg), &rawArr) == nil {
			for _, r := range rawArr {
				imageJSONs = append(imageJSONs, string(r))
			}
		}
	}
	for len(imageJSONs) < len(goodsIDs) {
		imageJSONs = append(imageJSONs, "")
	}
	if err := service.CreateOrderReviewsShopXO(c.GetUint("user_id"), uint(orderID), goodsIDs, ratings, contents, imageJSONs); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, map[string]interface{}{"msg": "评价成功"})
}

func sxCustomviewIndex(c *gin.Context) {
	id, _ := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if id == 0 {
		id, _ = strconv.ParseUint(c.Query("id"), 10, 64)
	}
	var v model.CustomView
	if err := app.Must().DB.Where("id = ? AND status = 1", id).First(&v).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "页面不存在")
		return
	}
	page := map[string]interface{}{
		"id": v.ID, "name": v.Title, "logo": "", "seo_title": v.Title, "seo_desc": "",
		"data": map[string]interface{}{"content": v.Content},
	}
	response.OK(c, map[string]interface{}{"data": page})
}

func sxDesignIndex(c *gin.Context) {
	id, _ := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if id == 0 {
		id, _ = strconv.ParseUint(c.Query("id"), 10, 64)
	}
	var d model.Design
	if err := app.Must().DB.Where("id = ? AND status = 1", id).First(&d).Error; err != nil {
		response.OK(c, map[string]interface{}{
			"data": []interface{}{}, "layout_data": []interface{}{}, "is_result_data_cache": 0,
		})
		return
	}
	var raw map[string]interface{}
	_ = json.Unmarshal([]byte(d.Data), &raw)
	widgets, _ := raw["data"].([]interface{})
	if widgets == nil {
		widgets = []interface{}{}
	}
	layout, _ := raw["layout_data"].([]interface{})
	if layout == nil {
		layout = []interface{}{}
	}
	response.OK(c, map[string]interface{}{
		"id": d.ID, "name": d.Name, "seo_title": d.Name, "seo_desc": "",
		"data": widgets, "layout_data": layout, "is_result_data_cache": 0,
	})
}

func sxFormInputVerifyEntry(c *gin.Context) {
	handler.ShopXOCompatCaptchaPNG(c)
}

func sxFormInputDataDetail(c *gin.Context) {
	id := getID(c)
	m, err := service.FormInputDataDetailForUser(c.GetUint("user_id"), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, m)
}

func sxUeditorIndex(c *gin.Context) {
	if c.PostForm("action") != "uploadimage" {
		response.OK(c, map[string]interface{}{"state": "SUCCESS"})
		return
	}
	file, err := c.FormFile("upfile")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "请选择文件")
		return
	}
	if file.Size > 5*1024*1024 {
		response.Fail(c, http.StatusBadRequest, "文件不能超过5MB")
		return
	}
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		response.Fail(c, http.StatusBadRequest, "不支持的文件格式")
		return
	}
	dir := fmt.Sprintf("uploads/%s", time.Now().Format("2006/01/02"))
	_ = os.MkdirAll(dir, 0755)
	dst := filepath.Join(dir, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Fail(c, http.StatusInternalServerError, "上传失败")
		return
	}
	response.OK(c, gin.H{"url": "/" + dst})
}

func sxUserVerifyEntry(c *gin.Context) {
	handler.ShopXOCompatCaptchaPNG(c)
}
