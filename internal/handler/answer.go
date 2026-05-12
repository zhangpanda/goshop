package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func CreateAnswer(c *gin.Context) {
	var req service.AnswerCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	a, err := service.CreateAnswer(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, a)
}

func GetAnswerList(c *gin.Context) {
	goodsID, _ := strconv.ParseUint(c.Query("goods_id"), 10, 64)
	page, pageSize := QueryPage(c)
	list, total, err := service.GetAnswerList(uint(goodsID), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func AdminReplyAnswer(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	service.ReplyAnswer(uint(id), req.Reply)
	response.OK(c, nil)
}

func AdminDeleteAnswer(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	service.DeleteAnswer(uint(id))
	response.OK(c, nil)
}
