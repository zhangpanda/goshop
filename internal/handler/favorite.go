package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/service"
	"github.com/zhangpanda/goshop/pkg/response"
)

func ToggleFavorite(c *gin.Context) {
	id, ok := ParamID(c, "id")
	if !ok {
		return
	}
	added, err := service.ToggleFavorite(c.GetUint("user_id"), uint(id))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"is_favorite": added})
}

func GetFavorites(c *gin.Context) {
	page, pageSize := QueryPage(c)
	list, total, err := service.GetFavorites(c.GetUint("user_id"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func GetBrowseHistory(c *gin.Context) {
	page, pageSize := QueryPage(c)
	list, total, err := service.GetBrowseHistory(c.GetUint("user_id"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": list})
}

func ClearBrowseHistory(c *gin.Context) {
	if err := service.ClearBrowseHistory(c.GetUint("user_id")); err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, nil)
}
