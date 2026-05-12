package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func CreateSeckill(c *gin.Context) {
	var req service.SeckillReq
	if !BindJSON(c, &req) {
		return
	}
	promo, err := service.CreateSeckill(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, promo)
}

func GetSeckillList(c *gin.Context) {
	page, pageSize := QueryPage(c)
	total, list, err := service.GetSeckillList(page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func GetActiveSeckills(c *gin.Context) {
	list, err := service.GetActiveSeckills()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func SeckillBuy(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil || itemID == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的商品ID")
		return
	}
	if err := service.SeckillBuy(c.GetUint("user_id"), uint(itemID)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
