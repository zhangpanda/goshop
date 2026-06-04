package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/pkg/response"
)

var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// allowedMIME 基于 magic bytes 检测的合法 MIME 类型
var allowedMIME = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
}

// dangerousPatterns 文件内容中不应出现的恶意代码特征
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<\?php`),
	regexp.MustCompile(`(?i)<script\b`),
	regexp.MustCompile(`(?i)<[a-z0-9_-]+:script\b`),
	regexp.MustCompile(`(?i)<iframe\b`),
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

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		response.Fail(c, http.StatusBadRequest, "不支持的文件格式")
		return
	}

	// 打开文件读取前512字节用于 MIME 检测和内容安全扫描
	src, err := file.Open()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "文件读取失败")
		return
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	buf = buf[:n]

	// MIME type 校验（基于 magic bytes）
	mimeType := http.DetectContentType(buf)
	if !allowedMIME[mimeType] {
		response.Fail(c, http.StatusBadRequest, "文件内容与类型不匹配")
		return
	}

	// 文件内容安全扫描（检测恶意代码注入）
	content := string(buf)
	for _, pat := range dangerousPatterns {
		if pat.MatchString(content) {
			response.Fail(c, http.StatusBadRequest, "文件包含非法内容")
			return
		}
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
