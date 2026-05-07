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

**核心原则：同域反代，不跨域（推荐）。**

未配置 `server.cors_origins` 时，CORS 仅放行 `http://localhost:*` 与 `http://127.0.0.1:*`（开发）。生产环境**推荐**反向代理把前端与 API 收敛到**同一站点**，浏览器不触发 CORS。若必须前后端分离域名，再在 `config.yaml` 中配置 `server.cors_origins` 为**完整 Origin** 白名单（如 `https://shop.example.com`）。

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

`cmd/server/main.go` 根据 **`server.trusted_proxies`** 调用 `SetTrustedProxies`：

- **列表非空**：仅信任来自这些网段/地址的反代，Gin 可从 `X-Forwarded-For` / `X-Real-IP` 解析真实客户端 IP（**务必将列表收窄为实际反代来源**，避免任意客户端伪造 `X-Forwarded-For` 绕过 IP 限流）。
- **列表为空（默认）**：等价于不信任代理头，`ClientIP()` 为直连对端。GoShop 若在 Nginx 之后且未配置本项，**限流、日志中的 IP 会长期为 `127.0.0.1`**。

**生产环境在反代后部署时，建议在 config.yaml 中配置：**

```yaml
server:
  port: 8080
  mode: release
  trusted_proxies:
    - "127.0.0.1"
    - "10.0.0.0/8"
```

按实际反代所在网段调整（例如仅 `127.0.0.1` 若只有本机 Nginx）。

### CORS

| 场景 | 是否可用 | 说明 |
|------|----------|------|
| 同域反代（推荐） | ✅ | 浏览器不触发 CORS，`cors_origins` 可不配 |
| 前后端分离域名 | ✅（需配置） | 在 `server.cors_origins` 中列出允许的完整 Origin（与浏览器请求 `Origin` 头**完全一致**，含协议与端口） |

实现见 `internal/middleware/cors.go`（读取 `global.Cfg.Server.CorsOrigins`）。

### 限流

当前限流中间件基于**进程内内存**，特点：

| 部署方式 | 效果 |
|----------|------|
| 单实例 | 正常工作 |
| 多实例（无共享） | 每个实例独立计数，实际限流阈值 = 配置值 × 实例数 |
| 多实例 + Redis | 需替换为 Redis 实现（当前未内置） |

**建议：** 小规模部署（≤3 实例）可接受当前方案；大规模部署应在网关层（Nginx `limit_req` 或 API Gateway）做限流。

### SQL Console

SQL Console 接口（`POST /api/admin/sql-console`，管理端「工具」权限）默认**关闭**。启用需在 config.yaml 中显式配置：

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
| `GOSHOP_SKIP_DEFAULT_ADMIN` | 为 `true` 时跳过创建默认管理员（适合自建账号流程） | （未设置） |
| `GOSHOP_E2E` | 为 `1` 时管理端登录跳过验证码（**仅 CI/本地 E2E，禁止生产**） | （未设置） |
| `NEXT_PUBLIC_BASE_PATH` | 前端 basePath | 空 |
| `NODE_ENV` | Next.js 环境 | `development` |

## 最小生产 checklist

- [ ] **⚠️ 首次部署后立即修改默认管理员密码**（默认 admin/admin123，公网暴露等于送入口）
- [ ] Nginx 反代配置完成（同域）
- [ ] `config.yaml` 中 `server.mode: release`
- [ ] `server.sql_console: false`（或不配置，默认关闭）
- [ ] `server.trusted_proxies` 配置为实际反代来源（若在 Nginx/K8s Ingress 后）
- [ ] 若浏览器跨域调 API：已配置 `server.cors_origins`，且与同域反代方案二选一论证过
- [ ] JWT Secret 使用强随机值（≥32 字符）
- [ ] 数据库账号最小权限（应用账号不给 DROP/ALTER）
- [ ] Redis 配置密码（如使用）
- [ ] 微信/支付宝密钥通过文件或环境变量注入，不入 git
