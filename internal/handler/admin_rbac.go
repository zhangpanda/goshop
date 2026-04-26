package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

// ========== 管理员登录 ==========

func AdminLoginHandler(c *gin.Context) {
	var req service.AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	// 验证码校验
	if req.CaptchaKey == "" || req.CaptchaCode == "" {
		response.Fail(c, http.StatusBadRequest, "请输入验证码")
		return
	}
	stored, err := global.RDB.Get(c, req.CaptchaKey).Result()
	if err != nil || stored != req.CaptchaCode {
		response.Fail(c, http.StatusBadRequest, "验证码错误")
		return
	}
	global.RDB.Del(c, req.CaptchaKey) // 用后即删

	resp, err := service.AdminLogin(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

func CreateAdminHandler(c *gin.Context) {
	var req service.CreateAdminReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	admin, err := service.CreateAdmin(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, admin)
}

func GetAdminListHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	total, list, err := service.GetAdminList(page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func UpdateAdminStatusHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.UpdateAdminStatusReq
	c.ShouldBindJSON(&req)
	if err := service.UpdateAdminStatus(uint(id), req.Status); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

// ========== 角色管理 ==========

func CreateRoleHandler(c *gin.Context) {
	var req service.RoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	role, err := service.CreateRole(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, role)
}

func GetRoleListHandler(c *gin.Context) {
	list, err := service.GetRoleList()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func UpdateRoleHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.RoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.UpdateRole(uint(id), &req); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}

func DeleteRoleHandler(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteRole(uint(id)); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, nil)
}
