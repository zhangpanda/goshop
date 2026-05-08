package handler

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	mrand "math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/pkg/response"
)

// secureDigits 返回一个 crypto/rand 生成的 n 位数字字符串（防预测）。
func secureDigits(n int) string {
	if n <= 0 {
		return ""
	}
	// 6 位数字最大 999999，用 uint32 足够；为安全余量分配 8 字节
	var buf [8]byte
	_, _ = crand.Read(buf[:])
	v := binary.BigEndian.Uint64(buf[:])
	// 取模到 10^n（最多 19 位内安全）
	mod := uint64(1)
	for i := 0; i < n; i++ {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", n, v%mod)
}

// checkCaptchaRate 使用 Cache.Incr + Expire 做原子限速：
// 同一 IP 60s 内超过 5 次生成请求则拒绝，避免原 Get→Set 方式下的 TOCTOU 竞态。
func checkCaptchaRate(c *gin.Context, ip string) bool {
	key := fmt.Sprintf("captcha_rate:%s", ip)
	n, err := app.Must().Cache.Incr(c, key)
	if err != nil {
		// 限速系统异常时放行，不让它变成拒绝服务向量
		return true
	}
	if n == 1 {
		_ = app.Must().Cache.Expire(c, key, 60*time.Second)
	}
	return n <= 5
}

// AdminCaptcha 生成图片验证码
func AdminCaptcha(c *gin.Context) {
	if !checkCaptchaRate(c, c.ClientIP()) {
		response.Fail(c, 429, "请求过于频繁，请稍后再试")
		return
	}
	code := secureDigits(6)
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
	code := secureDigits(6)
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
	// 图像装饰噪点/颜色抖动只影响可读性，不是安全面，沿用 math/rand 即可。
	for i := 0; i < 150; i++ {
		img.Set(mrand.Intn(160), mrand.Intn(40), color.RGBA{uint8(mrand.Intn(200)), uint8(mrand.Intn(200)), uint8(mrand.Intn(200)), 255})
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
	c := color.RGBA{uint8(mrand.Intn(100)), uint8(mrand.Intn(100)), uint8(mrand.Intn(150)), 255}
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
