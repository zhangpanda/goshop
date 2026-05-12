package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func Register(c *gin.Context) {
	var req service.RegisterReq
	if !BindJSON(c, &req) {
		return
	}
	user, err := service.Register(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, user)
}

func Login(c *gin.Context) {
	var req service.LoginReq
	if !BindJSON(c, &req) {
		return
	}
	resp, err := service.Login(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}

func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	user, err := service.GetUserByID(userID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	response.OK(c, user)
}
