package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/pkg/response"
)

// BindJSON 绑定 JSON body 并校验。失败时自动返回 400 并 Abort。
// 返回 true 表示绑定成功，false 表示已响应错误。
func BindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		c.Abort()
		return false
	}
	return true
}

// BindQuery 绑定 query string 并校验。
func BindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		c.Abort()
		return false
	}
	return true
}

// ParamID 从 URL 路径参数解析 uint ID。失败时自动返回 400。
func ParamID(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的ID参数")
		c.Abort()
		return 0, false
	}
	return uint(id), true
}

// QueryPage 解析分页参数，返回 page 和 pageSize（带默认值）。
func QueryPage(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return
}
