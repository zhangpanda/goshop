package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func CreateGroupBuy(c *gin.Context) {
	var req service.GroupBuyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	promo, err := service.CreateGroupBuy(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, promo)
}

func GetGroupBuyList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	total, list, err := service.GetGroupBuyList(page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func GetActiveGroupBuys(c *gin.Context) {
	list, err := service.GetActiveGroupBuys()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func OpenGroup(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil || itemID == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的商品ID")
		return
	}
	grp, err := service.OpenGroup(c.GetUint("user_id"), uint(itemID))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, grp)
}

func JoinGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || groupID == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的拼团ID")
		return
	}
	if err := service.JoinGroup(c.GetUint("user_id"), uint(groupID)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func GetGroupOrderDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的ID")
		return
	}
	grp, members, err := service.GetGroupOrderDetail(uint(id))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"group": grp, "members": members})
}
