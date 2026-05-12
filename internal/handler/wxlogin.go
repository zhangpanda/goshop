package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func WxLogin(c *gin.Context) {
	var req service.WxLoginReq
	if !BindJSON(c, &req) {
		return
	}
	resp, err := service.WxLogin(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, resp)
}
