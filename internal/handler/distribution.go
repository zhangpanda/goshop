package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 用户端 ==========

func ApplyDistributor(c *gin.Context) {
	var req struct {
		ParentID uint `json:"parent_id"`
	}
	if !BindJSON(c, &req) {
		return
	}
	d, err := service.ApplyDistributor(c.GetUint("user_id"), req.ParentID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, d)
}

func GetMyDistributor(c *gin.Context) {
	d, err := service.GetDistributorByUser(c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, d)
}

func GetMyTeam(c *gin.Context) {
	list, err := service.GetSubDistributors(c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func GetMyCommissionLogs(c *gin.Context) {
	d, err := service.GetDistributorByUser(c.GetUint("user_id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	page, pageSize := QueryPage(c)
	total, list, _ := service.GetCommissionLogs(d.ID, page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

func RequestWithdraw(c *gin.Context) {
	var req struct {
		Amount      int64  `json:"amount" binding:"required,min=1"`
		AccountType string `json:"account_type" binding:"required"`
		AccountNo   string `json:"account_no" binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.RequestWithdraw(c.GetUint("user_id"), req.Amount, req.AccountType, req.AccountNo, req.AccountName); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

// ========== 管理端 ==========

func AdminCreateDistributor(c *gin.Context) {
	var req struct {
		UserID   uint `json:"user_id" binding:"required"`
		ParentID uint `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	d, err := service.ApplyDistributor(req.UserID, req.ParentID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, d)
}

func AdminDistributorList(c *gin.Context) {
	page, pageSize := QueryPage(c)
	total, list, _ := service.GetDistributorList(page, pageSize)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminWithdrawList(c *gin.Context) {
	page, pageSize := QueryPage(c)
	total, list, _ := service.GetWithdrawList(page, pageSize, nil)
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminAuditWithdraw(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "无效ID")
		return
	}
	var req struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}
	if !BindJSON(c, &req) {
		return
	}
	if err := service.AuditWithdraw(uint(id), req.Approve, req.Reason); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
