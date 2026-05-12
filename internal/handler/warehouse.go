package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func CreateWarehouse(c *gin.Context) {
	var req service.WarehouseReq
	if !BindJSON(c, &req) {
		return
	}
	w, err := service.CreateWarehouse(&req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, w)
}

func GetWarehouseList(c *gin.Context) {
	list, _ := service.GetWarehouseList()
	response.OK(c, list)
}

func UpdateWarehouse(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var req service.WarehouseReq
	if !BindJSON(c, &req) {
		return
	}
	if err := service.UpdateWarehouse(uint(id), &req); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

func DeleteWarehouse(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteWarehouse(uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func WarehouseGoodsAdd(c *gin.Context) {
	var req struct {
		WarehouseID uint `json:"warehouse_id" binding:"required"`
		GoodsID     uint `json:"goods_id" binding:"required"`
		Inventory   int  `json:"inventory"`
	}
	if !BindJSON(c, &req) {
		return
	}
	service.WarehouseGoodsAdd(req.WarehouseID, req.GoodsID, req.Inventory)
	response.OK(c, nil)
}

func WarehouseGoodsList(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	list, _ := service.WarehouseGoodsList(uint(id))
	response.OK(c, list)
}

func WarehouseGoodsSpecSave(c *gin.Context) {
	var req struct {
		WarehouseID uint   `json:"warehouse_id" binding:"required"`
		GoodsID     uint   `json:"goods_id" binding:"required"`
		SKUID       uint   `json:"sku_id" binding:"required"`
		Inventory   int    `json:"inventory"`
		SpecValues  string `json:"spec_values"`
	}
	if !BindJSON(c, &req) {
		return
	}
	service.WarehouseGoodsSpecSave(req.WarehouseID, req.GoodsID, req.SKUID, req.Inventory, req.SpecValues)
	response.OK(c, nil)
}
