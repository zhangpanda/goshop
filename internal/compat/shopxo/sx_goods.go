package shopxo

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func sxGoodsDetail(c *gin.Context) {
	id := getID(c)
	goods, err := service.GetGoodsDetail(id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "商品不存在")
		return
	}
	response.OK(c, goods)
}

func sxGoodsCategory(c *gin.Context) {
	cats := service.GoodsCategoryAll()
	response.OK(c, cats)
}

func sxGoodsFavor(c *gin.Context) {
	id := getID(c)
	added, _ := service.ToggleFavorite(c.GetUint("user_id"), id)
	response.OK(c, map[string]interface{}{"is_favorite": added})
}

func sxGoodsSpecType(c *gin.Context) {
	id := getID(c)
	response.OK(c, service.GoodsSpecType(id))
}

func sxGoodsSpecDetail(c *gin.Context) {
	id := getID(c)
	spec := c.Query("spec")
	resp, err := service.GoodsSpecDetail(id, spec)
	if err != nil {
		response.OK(c, map[string]interface{}{"spec": nil, "sku": nil})
		return
	}
	response.OK(c, resp)
}

func sxGoodsStock(c *gin.Context) {
	id := getID(c)
	stock, _ := service.GoodsStock(id, c.Query("spec"))
	response.OK(c, map[string]int{"stock": stock})
}

func sxGoodsScore(c *gin.Context) {
	response.OK(c, service.GoodsScore(getID(c)))
}

func sxGoodsComments(c *gin.Context) {
	id := getID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	list, total, _ := service.GetGoodsReviews(id, page, 10)
	response.OK(c, map[string]interface{}{"total": total, "data": list})
}

func sxSearchIndex(c *gin.Context) {
	var req service.GoodsListReq
	c.ShouldBindQuery(&req)
	if wd := c.Query("wd"); wd != "" {
		req.Keyword = wd
	}
	if kw := c.Query("keywords"); kw != "" {
		req.Keyword = kw
	}
	resp, _ := service.GetGoodsList(&req)
	response.OK(c, resp)
}
func sxSearchDataList(c *gin.Context) {
	var req service.GoodsListReq
	c.ShouldBindQuery(&req)
	if wd := c.Query("wd"); wd != "" {
		req.Keyword = wd
	}
	resp, _ := service.GetGoodsList(&req)
	response.OK(c, resp)
}
func sxSearchStart(c *gin.Context) { response.OK(c, service.SearchStartData()) }

