package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func CreateReview(c *gin.Context) {
	var req service.CreateReviewReq
	if !BindJSON(c, &req) {
		return
	}
	review, err := service.CreateReview(c.GetUint("user_id"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, review)
}

func GetGoodsReviews(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	page, pageSize := QueryPage(c)
	list, total, err := service.GetGoodsReviews(uint(id), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func ReplyReview(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	var req struct {
		Reply string `json:"reply" binding:"required"`
	}
	if !BindJSON(c, &req) {
		return
	}
	if err := service.ReplyReview(uint(id), req.Reply); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}
