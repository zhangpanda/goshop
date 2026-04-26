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
	var req struct { Name string `json:"name" binding:"required"`; Code string `json:"code" binding:"required"`; Icon string `json:"icon"`; Sort int `json:"sort"` }
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	e, _ := service.CreateExpress(req.Name, req.Code, req.Icon, req.Sort)
	response.OK(c, e)
}
func GetExpressList(c *gin.Context) { list, _ := service.GetExpressList(); response.OK(c, list) }

// 库存日志
func GetInventoryLogList(c *gin.Context) {
	gid, _ := strconv.ParseUint(c.Query("goods_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, _ := service.GetInventoryLogList(uint(gid), page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

// 订单删除
func DeleteOrderHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteOrder(c.GetUint("user_id"), uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, "只能删除已完成或已取消的订单"); return
	}
	response.OK(c, nil)
}

// 权限树
func CreatePowerHandler(c *gin.Context) {
	var req struct { ParentID uint `json:"parent_id"`; Name string `json:"name" binding:"required"`; Control string `json:"control"`; Sort int `json:"sort"` }
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	p, _ := service.CreatePower(req.ParentID, req.Name, req.Control, req.Sort)
	response.OK(c, p)
}
func GetPowerTree(c *gin.Context) { list, _ := service.GetPowerTree(); response.OK(c, list) }
func SaveRolePowers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct { PowerIDs []uint `json:"power_ids"` }
	c.ShouldBindJSON(&req)
	service.SaveRolePowers(uint(id), req.PowerIDs)
	response.OK(c, nil)
}
func GetRolePowersHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ids, _ := service.GetRolePowers(uint(id))
	response.OK(c, ids)
}

// 品牌分类
func CreateBrandCategory(c *gin.Context) {
	var req struct { Name string `json:"name" binding:"required"`; Sort int `json:"sort"` }
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	bc := service.CreateBrandCategoryRecord(req.Name, req.Sort)
	response.OK(c, bc)
}
func GetBrandCategoryList(c *gin.Context) { list := service.GetBrandCategoryListRecords(); response.OK(c, list) }

// 商品多分类
func SaveGoodsCategoryJoin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct { CategoryIDs []uint `json:"category_ids"` }
	c.ShouldBindJSON(&req)
	service.SaveGoodsCategoryJoinRecords(uint(id), req.CategoryIDs)
	response.OK(c, nil)
}
