package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func SendVerifyCode(c *gin.Context) {
	var req struct {
		Account string `json:"account" binding:"required"`
		Type    string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.SendVerifyCode(req.Account, req.Type)
	response.OK(c, nil)
}

func UpdatePassword(c *gin.Context) {
	var req service.UpdatePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.UpdatePassword(c.GetUint("user_id"), &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func ForgetPassword(c *gin.Context) {
	var req service.ForgetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.ForgetPassword(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func BindMobile(c *gin.Context) {
	var req service.BindMobileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.BindMobile(c.GetUint("user_id"), &req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}

func GetUserPlatforms(c *gin.Context) {
	list, _ := service.GetUserPlatforms(c.GetUint("user_id"))
	response.OK(c, list)
}
