package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/config"
	"github.com/zhangpanda/goshop/internal/app"
	"github.com/zhangpanda/goshop/internal/model"
	"github.com/zhangpanda/goshop/internal/router"
	"github.com/zhangpanda/goshop/internal/testutil"
)

// TestMain 搭建 handler 测试共用的依赖：
//   - testutil.SetupTestDB()：in-memory SQLite + AutoMigrate + repository.Init
//   - 用 app.Must() 拿到的 Deps 上补 Cfg（router.Setup 读 MetricsPath / JWT 等）
//   - 小量 seed（站点开关、登录必需的配置项）
//
// 所有 handler 包测试共用一个 SQLite 库；需要真 MySQL 的测试请在单测内部
// 调用 testutil.SetupMySQLAppDeps(t)。
func TestMain(m *testing.M) {
	testutil.SetupTestDB()
	// Cfg 字段补全：router.Setup 读 MetricsPath / RateLimit / JWT
	app.Must().Cfg = &config.Config{
		Server: config.ServerConfig{MetricsPath: ""},
		JWT:    config.JWTConfig{Secret: "test-secret-at-least-32-chars-xxxxx", Expire: 24},
	}
	// 最小站点配置，供 /api/site-config 等读取
	app.Must().DB.Create(&model.Config{Key: "site_name", Value: "goshop-test"})
	os.Exit(m.Run())
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	router.Setup(r)
	return r
}

type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	r.ServeHTTP(w, req)
	return w
}

func doPost(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func parseResp(w *httptest.ResponseRecorder) apiResp {
	var resp apiResp
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// TestPublicAPIs 测试公共只读接口：HTTP 200 即视为通过，内部 code 容忍 0/-1（无数据返回空列表也算）。
func TestPublicAPIs(t *testing.T) {
	r := setupRouter()

	tests := []struct {
		name string
		path string
	}{
		{"商品列表", "/api/goods"},
		{"分类树", "/api/categories"},
		{"文章列表", "/api/articles"},
		{"文章分类", "/api/article-categories"},
		{"优惠券列表", "/api/coupons"},
		{"站点配置", "/api/site-config"},
		{"轮播图", "/api/slides"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doGet(r, tt.path)
			if w.Code != http.StatusOK {
				t.Errorf("GET %s: status=%d, want 200", tt.path, w.Code)
			}
		})
	}
}

// TestAuthRequired 已登录才可访问的接口，匿名访问必须失败（Code != 0 或 HTTP 401）。
func TestAuthRequired(t *testing.T) {
	r := setupRouter()

	paths := []string{"/api/cart", "/api/orders", "/api/address", "/api/favorites", "/api/user/profile"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			w := doGet(r, path)
			if w.Code == http.StatusOK {
				resp := parseResp(w)
				if resp.Code == 0 {
					t.Errorf("GET %s without token should fail, got code=0 body=%s", path, w.Body.String())
				}
			}
		})
	}
}

// TestRegisterLogin 完整的注册 → 登录闭环。验证密码 hash、token 签发、user 持久化。
func TestRegisterLogin(t *testing.T) {
	r := setupRouter()
	// 避免 TestMain seed 外其他用例污染：用独立 username
	body := `{"username":"bot_reglogin","password":"test123456"}`

	// 注册
	w := doPost(r, "/api/register", body)
	if w.Code != http.StatusOK {
		t.Fatalf("register http status=%d body=%s", w.Code, w.Body.String())
	}
	resp := parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("register failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	// 登录
	w = doPost(r, "/api/login", body)
	if w.Code != http.StatusOK {
		t.Fatalf("login http status=%d body=%s", w.Code, w.Body.String())
	}
	resp = parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("login failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	var loginData struct {
		Token string `json:"token"`
	}
	json.Unmarshal(resp.Data, &loginData)
	if loginData.Token == "" {
		t.Fatal("login returned empty token")
	}
	// 简单 sanity：JWT 三段式
	parts := strings.Split(loginData.Token, ".")
	if len(parts) != 3 {
		t.Errorf("token not a well-formed JWT: %s", loginData.Token)
	}
}
