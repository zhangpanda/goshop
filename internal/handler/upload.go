package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/pkg/response"
)

var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "请选择文件")
		return
	}
	if file.Size > 5*1024*1024 {
		response.Fail(c, http.StatusBadRequest, "文件不能超过5MB")
		return
	}

	ext := filepath.Ext(file.Filename)
	if !allowedExts[ext] {
		response.Fail(c, http.StatusBadRequest, "不支持的文件格式")
		return
	}

	// 按日期分目录
	dir := fmt.Sprintf("uploads/%s", time.Now().Format("2006/01/02"))
	os.MkdirAll(dir, 0755)

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Fail(c, http.StatusInternalServerError, "上传失败")
		return
	}

	response.OK(c, gin.H{"url": "/" + dst})
}
