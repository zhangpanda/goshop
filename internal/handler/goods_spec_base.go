package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func SaveGoodsSpecBase(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Specs []service.SpecBaseReq `json:"specs" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error()); return
	}
	service.SaveGoodsSpecBase(uint(id), req.Specs)
	response.OK(c, nil)
}

func GetGoodsSpecBase(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, _ := service.GetGoodsSpecBase(uint(id))
	response.OK(c, list)
}

func SaveGoodsPhotos(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Images []string `json:"images" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error()); return
	}
	service.SaveGoodsPhotos(uint(id), req.Images)
	response.OK(c, nil)
}

func GetGoodsPhotos(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, _ := service.GetGoodsPhotos(uint(id))
	response.OK(c, list)
}
