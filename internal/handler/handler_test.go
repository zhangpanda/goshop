package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zhangpanda/goshop/internal/router"
)

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

// TestPublicAPIs 测试不需要登录的公共接口
func TestPublicAPIs(t *testing.T) {
	// 注意：这些测试需要数据库连接，属于集成测试
	// 在 CI 中需要先启动 MySQL 和 Redis
	t.Skip("需要数据库连接，跳过集成测试")

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
			if w.Code != 200 {
				t.Errorf("GET %s: status=%d, want 200", tt.path, w.Code)
			}
			resp := parseResp(w)
			if resp.Code != 0 {
				t.Errorf("GET %s: code=%d msg=%s, want code=0", tt.path, resp.Code, resp.Msg)
			}
		})
	}
}

// TestAuthRequired 测试需要登录的接口返回 401
func TestAuthRequired(t *testing.T) {
	t.Skip("需要数据库连接，跳过集成测试")

	r := setupRouter()

	paths := []string{"/api/cart", "/api/orders", "/api/address", "/api/favorites", "/api/user/profile"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			w := doGet(r, path)
			resp := parseResp(w)
			if resp.Code == 0 {
				t.Errorf("GET %s without token should fail, got code=0", path)
			}
		})
	}
}

// TestRegisterLogin 测试注册登录流程
func TestRegisterLogin(t *testing.T) {
	t.Skip("需要数据库连接，跳过集成测试")

	r := setupRouter()

	// 注册
	w := doPost(r, "/api/register", `{"username":"testbot","password":"test123456"}`)
	resp := parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("register failed: %s", resp.Msg)
	}

	// 登录
	w = doPost(r, "/api/login", `{"username":"testbot","password":"test123456"}`)
	resp = parseResp(w)
	if resp.Code != 0 {
		t.Fatalf("login failed: %s", resp.Msg)
	}

	var loginData struct{ Token string `json:"token"` }
	json.Unmarshal(resp.Data, &loginData)
	if loginData.Token == "" {
		t.Fatal("login returned empty token")
	}
}
