package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ---- 分类 ----

func CreateCategory(c *gin.Context) {
	var req service.CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	cat, err := service.CreateCategory(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, cat)
}

func GetCategoryTree(c *gin.Context) {
	cats, err := service.GetCategoryTree()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, cats)
}

// ---- 商品 ----

func CreateGoods(c *gin.Context) {
	var req service.CreateGoodsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	goods, err := service.CreateGoods(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, goods)
}

func GetGoodsList(c *gin.Context) {
	var req service.GoodsListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := service.GetGoodsList(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, resp)
}

func GetGoodsDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "无效的商品ID")
		return
	}
	goods, err := service.GetGoodsDetail(uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "商品不存在")
		return
	}
	response.OK(c, goods)
}
