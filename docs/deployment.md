# 部署指南

## 推荐部署拓扑

```
                    ┌─────────────┐
  浏览器/App ──────▶│  Nginx/网关  │
                    │  (443/80)   │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
        /goshop/*    /goshop/admin/*   /api/*  /uploads/*
              │            │            │
              ▼            ▼            ▼
         web (3000)  admin (3001)   GoShop (8080)
```

**核心原则：同域反代，不跨域。**

GoShop 的 CORS 策略仅放行 `localhost/127.0.0.1`（开发用途）。生产环境必须通过反向代理将前端和 API 统一到同一域名下，浏览器不会触发跨域请求。

## Nginx 配置示例

```nginx
server {
    listen 443 ssl;
    server_name shop.example.com;

    # 前台
    location /goshop/ {
        proxy_pass http://127.0.0.1:3000/;
    }

    # 管理端
    location /goshop/admin/ {
        proxy_pass http://127.0.0.1:3001/;
    }

    # API + 静态资源
    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
    }

    # ShopXO uni-app 兼容入口
    location = /api.php {
        proxy_pass http://127.0.0.1:8080/api.php;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 安全相关配置

### TrustedProxies

当前代码 `r.SetTrustedProxies(nil)` 表示不信任任何代理头。这意味着：

- `c.ClientIP()` 返回的是 TCP 连接的直接对端 IP
- 如果 GoShop 在 Nginx 后面，`ClientIP()` 永远是 `127.0.0.1`
- **限流、风控、日志中的 IP 都会失效**

**生产环境必须配置 TrustedProxies：**

```go
// main.go 中修改为：
r.SetTrustedProxies([]string{"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12"})
```

或通过 config.yaml 配置：

```yaml
server:
  port: 8080
  mode: release
  trusted_proxies:
    - "127.0.0.1"
    - "10.0.0.0/8"
```

配置后 Gin 会从 `X-Forwarded-For` / `X-Real-IP` 中提取真实客户端 IP。

### CORS

| 场景 | 是否可用 | 说明 |
|------|----------|------|
| 同域反代（推荐） | ✅ | 浏览器不触发 CORS，无需额外配置 |
| 前后端分离域名 | ❌ | 当前 CORS 策略会拒绝非 localhost origin |

如果必须跨域部署，需修改 `internal/middleware/cors.go` 中的 `AllowOriginFunc`，添加生产域名白名单。

### 限流

当前限流中间件基于**进程内内存**，特点：

| 部署方式 | 效果 |
|----------|------|
| 单实例 | 正常工作 |
| 多实例（无共享） | 每个实例独立计数，实际限流阈值 = 配置值 × 实例数 |
| 多实例 + Redis | 需替换为 Redis 实现（当前未内置） |

**建议：** 小规模部署（≤3 实例）可接受当前方案；大规模部署应在网关层（Nginx `limit_req` 或 API Gateway）做限流。

### SQL Console

SQL Console 功能（`/api/admin/sql`）默认**关闭**。启用需在 config.yaml 中显式配置：

```yaml
server:
  sql_console: true  # 仅开发/调试环境启用
```

即使启用，也仅限超级管理员、只读语句、10 秒超时。生产环境**强烈建议保持关闭**，通过数据库只读副本满足查询需求。

## AutoMigrate

`main.go` 中的 `AutoMigrate` 在每次启动时执行。生产环境风险：

- 大表加列可能锁表
- 意外删列（GORM 不会删，但第三方工具可能）

**生产建议：**

1. 首次部署后，将 AutoMigrate 移到独立的 migration 命令
2. 后续 schema 变更通过手工 DDL 或 migration 工具管理
3. 或通过环境变量控制：`GOSHOP_AUTO_MIGRATE=false` 跳过

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `GOSHOP_AUTO_MIGRATE` | 是否执行 AutoMigrate | `true` |
| `NEXT_PUBLIC_BASE_PATH` | 前端 basePath | 空 |
| `NODE_ENV` | Next.js 环境 | `development` |

## 最小生产 checklist

- [ ] Nginx 反代配置完成（同域）
- [ ] `config.yaml` 中 `server.mode: release`
- [ ] `server.sql_console: false`（或不配置，默认关闭）
- [ ] TrustedProxies 配置为实际代理 IP
- [ ] JWT Secret 使用强随机值（≥32 字符）
- [ ] 数据库账号最小权限（应用账号不给 DROP/ALTER）
- [ ] Redis 配置密码（如使用）
- [ ] 微信/支付宝密钥通过文件或环境变量注入，不入 git
