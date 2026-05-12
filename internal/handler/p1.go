package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// 快递公司
func CreateExpressHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
		Icon string `json:"icon"`
		Sort int    `json:"sort"`
	}
	if !BindJSON(c, &req) {
		return
	}
	e, _ := service.CreateExpress(req.Name, req.Code, req.Icon, req.Sort)
	response.OK(c, e)
}
func GetExpressList(c *gin.Context) { list, _ := service.GetExpressList(); response.OK(c, list) }

// 库存日志
func GetInventoryLogList(c *gin.Context) {
	gid, _ := strconv.ParseUint(c.Query("goods_id"), 10, 64)
	page, pageSize := QueryPage(c)
	list, total, _ := service.GetInventoryLogList(uint(gid), page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

// 订单删除
func DeleteOrderHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	if err := service.DeleteOrder(c.GetUint("user_id"), uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, "只能删除已完成或已取消的订单")
		return
	}
	response.OK(c, nil)
}

// 权限树
func CreatePowerHandler(c *gin.Context) {
	var req struct {
		ParentID uint   `json:"parent_id"`
		Name     string `json:"name" binding:"required"`
		Control  string `json:"control"`
		Sort     int    `json:"sort"`
	}
	if !BindJSON(c, &req) {
		return
	}
	p, _ := service.CreatePower(req.ParentID, req.Name, req.Control, req.Sort)
	response.OK(c, p)
}
func GetPowerTree(c *gin.Context) { list, _ := service.GetPowerTree(); response.OK(c, list) }
func SaveRolePowers(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var req struct {
		PowerIDs []uint `json:"power_ids"`
	}
	if !BindJSON(c, &req) {
		return
	}
	service.SaveRolePowers(uint(id), req.PowerIDs)
	response.OK(c, nil)
}
func GetRolePowersHandler(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	ids, _ := service.GetRolePowers(uint(id))
	response.OK(c, ids)
}

// 品牌分类
func CreateBrandCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Sort int    `json:"sort"`
	}
	if !BindJSON(c, &req) {
		return
	}
	bc := service.CreateBrandCategoryRecord(req.Name, req.Sort)
	response.OK(c, bc)
}
func GetBrandCategoryList(c *gin.Context) {
	list := service.GetBrandCategoryListRecords()
	response.OK(c, list)
}

// 商品多分类
func SaveGoodsCategoryJoin(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var req struct {
		CategoryIDs []uint `json:"category_ids"`
	}
	if !BindJSON(c, &req) {
		return
	}
	service.SaveGoodsCategoryJoinRecords(uint(id), req.CategoryIDs)
	response.OK(c, nil)
}
