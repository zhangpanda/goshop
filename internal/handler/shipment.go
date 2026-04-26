package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func ShipOrder(c *gin.Context) {
	var req service.ShipOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.ShipOrder(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func ConfirmReceive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.ConfirmReceive(c.GetUint("user_id"), uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func GetShipment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	s, err := service.GetShipment(uint(id))
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	// 查询物流轨迹
	tracks, _ := service.QueryExpress(s.ExpressCompany, s.ExpressNo)
	response.OK(c, gin.H{"shipment": s, "tracks": tracks})
}
