package handler

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/response"
)

// AdminCaptcha 生成图片验证码
func AdminCaptcha(c *gin.Context) {
	// 限速：同IP 60秒内最多5次
	ipKey := fmt.Sprintf("captcha_rate:%s", c.ClientIP())
	if count, _ := app.Must().Cache.Get(c, ipKey); count != "" {
		n := 0
		fmt.Sscanf(count, "%d", &n)
		if n >= 5 {
			response.Fail(c, 429, "请求过于频繁，请稍后再试")
			return
		}
		app.Must().Cache.Set(c, ipKey, fmt.Sprintf("%d", n+1), 60*time.Second)
	} else {
		app.Must().Cache.Set(c, ipKey, "1", 60*time.Second)
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	key := fmt.Sprintf("captcha:%s", c.Query("key"))
	if key == "captcha:" {
		key = fmt.Sprintf("captcha:%d", time.Now().UnixNano())
	}
	app.Must().Cache.Set(c, key, code, 5*time.Minute)
	writeCaptchaPNG(c, code, key)
}

// ShopXOCompatCaptchaPNG 供 shopxo-uniapp 在 /api.php 中通过 userverifyentry、forminput/verifyentry 拉取 PNG 验证码（<img src=...>）。
// 校验值存入 Cache，key 置于响应头 X-Captcha-Key；登录校验需客户端回传该 key（与 ShopXO 行为对齐程度取决于前端实现）。
func ShopXOCompatCaptchaPNG(c *gin.Context) {
	suffix := c.DefaultQuery("type", "")
	if suffix == "" {
		suffix = "form:" + c.DefaultQuery("t", "0")
	}
	key := fmt.Sprintf("captcha:shopxo:%s:%s", suffix, c.ClientIP())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	app.Must().Cache.Set(c, key, code, 5*time.Minute)
	writeCaptchaPNG(c, code, key)
}

func writeCaptchaPNG(c *gin.Context, code, cacheKey string) {
	img := image.NewRGBA(image.Rect(0, 0, 160, 40))
	for x := 0; x < 160; x++ {
		for y := 0; y < 40; y++ {
			img.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}
	for i := 0; i < 150; i++ {
		img.Set(rand.Intn(160), rand.Intn(40), color.RGBA{uint8(rand.Intn(200)), uint8(rand.Intn(200)), uint8(rand.Intn(200)), 255})
	}
	digits := []byte(code)
	for i, d := range digits {
		drawDigit(img, int(d-'0'), 15+i*25, 8)
	}
	c.Header("Content-Type", "image/png")
	c.Header("X-Captcha-Key", cacheKey)
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
	app.Must().DB.Model(&model.AdminOperationLog{}).Count(&total)
	var list []model.AdminOperationLog
	app.Must().DB.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.OK(c, gin.H{"total": total, "list": list})
}
