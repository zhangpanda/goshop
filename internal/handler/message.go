package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func GetMessages(c *gin.Context) {
	page, pageSize := QueryPage(c)
	list, total, err := service.GetMessages(c.GetUint("user_id"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	unread := service.UnreadCount(c.GetUint("user_id"))
	response.OK(c, gin.H{"total": total, "unread": unread, "list": list})
}

func ReadMessage(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	service.ReadMessage(c.GetUint("user_id"), uint(id))
	response.OK(c, nil)
}

func ReadAllMessages(c *gin.Context) {
	service.ReadAllMessages(c.GetUint("user_id"))
	response.OK(c, nil)
}
