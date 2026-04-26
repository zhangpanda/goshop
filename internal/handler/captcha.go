package handler

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/global"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/response"
)

// AdminCaptcha 生成图片验证码
func AdminCaptcha(c *gin.Context) {
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	// 存到 Redis，5分钟过期
	key := fmt.Sprintf("captcha:%s", c.Query("key"))
	if key == "captcha:" {
		key = fmt.Sprintf("captcha:%d", time.Now().UnixNano())
	}
	global.RDB.Set(c, key, code, 5*time.Minute)

	// 生成简单图片
	img := image.NewRGBA(image.Rect(0, 0, 120, 40))
	// 背景
	for x := 0; x < 120; x++ {
		for y := 0; y < 40; y++ {
			img.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}
	// 简单噪点
	for i := 0; i < 100; i++ {
		img.Set(rand.Intn(120), rand.Intn(40), color.RGBA{uint8(rand.Intn(200)), uint8(rand.Intn(200)), uint8(rand.Intn(200)), 255})
	}
	// 数字（简单像素字体）
	digits := []byte(code)
	for i, d := range digits {
		drawDigit(img, int(d-'0'), 15+i*25, 8)
	}

	c.Header("Content-Type", "image/png")
	c.Header("X-Captcha-Key", key)
	png.Encode(c.Writer, img)
}

func drawDigit(img *image.RGBA, digit, x, y int) {
	// 5x7 pixel font for digits 0-9
	fonts := [10][7]uint8{
		{0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E}, // 0
		{0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E}, // 1
		{0x0E, 0x11, 0x01, 0x06, 0x08, 0x10, 0x1F}, // 2
		{0x0E, 0x11, 0x01, 0x06, 0x01, 0x11, 0x0E}, // 3
		{0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02}, // 4
		{0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E}, // 5
		{0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E}, // 6
		{0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08}, // 7
		{0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E}, // 8
		{0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C}, // 9
	}
	c := color.RGBA{uint8(rand.Intn(100)), uint8(rand.Intn(100)), uint8(rand.Intn(150)), 255}
	for row := 0; row < 7; row++ {
		for col := 0; col < 5; col++ {
			if fonts[digit][row]&(1<<(4-col)) != 0 {
				for dx := 0; dx < 3; dx++ {
					for dy := 0; dy < 3; dy++ {
						img.Set(x+col*3+dx, y+row*3+dy, c)
					}
				}
			}
		}
	}
}

// AdminOperationLogList 操作日志列表
func AdminOperationLogList(c *gin.Context) {
	page := 1
	pageSize := 20
	fmt.Sscanf(c.DefaultQuery("page", "1"), "%d", &page)
	fmt.Sscanf(c.DefaultQuery("page_size", "20"), "%d", &pageSize)
	var total int64
	global.DB.Model(&model.AdminOperationLog{}).Count(&total)
	var list []model.AdminOperationLog
	global.DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
