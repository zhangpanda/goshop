# GoShop 二次开发指南

## 环境要求

与 **`go.mod`** 的 **`go major.minor.patch`**（**唯一真值**；`toolchain` 行可选，`go mod tidy` 常会删除与 `go` 重复的 `toolchain`）、**CI**、**`Dockerfile`**、**`mise.toml`**、**`.tool-versions`** 保持一致：

- **Go 1.25.11**（`python3 scripts/sync_go_toolchain.py` **无参**即可自检；与 `govulncheck` 等补丁要求冲突时请升级）
- **本机版本管理（可选）**：[mise](https://mise.jdx.dev/) 在仓库根执行 `mise install`（读 `mise.toml`）；[asdf](https://asdf-vm.com/) 执行 `asdf install`（读 `.tool-versions` 的 `golang`）
- **Node.js 20+**（Next.js 15）
- **MySQL 5.7+ / 8.0+**（推荐 8.0）
- **Redis 6+**（可选；不配则用内存缓存）

### 升级 Go 补丁版本

1. 只改 **`go.mod`** 的 **`go major.minor.patch`** 一行（例如升到 `1.25.11`）。
2. 执行 **`python3 scripts/sync_go_toolchain.py --write`**，再 **`go mod tidy`**。脚本会回写 **`toolchain`（若 `tidy` 删了可再跑一次 `--write` 补回，纯属可选）**、`Dockerfile`、`ci.yml`、`mise.toml`、`.tool-versions` 及常见 Markdown。**CI 在 `setup-go` 前跑本脚本**，分叉直接失败。
3. **不要**在文档里手写与 `go.mod` 不一致的示例版本号。

## 项目架构

```
goshop/
├── cmd/server/main.go              # 入口：加载配置 → 初始化DB/Redis → app.Register → 可选建表 → Seed → 启动HTTP
├── cmd/migrate/main.go             # 独立迁移：RunAllSchemaMigrations（SQL 版本 + 可选 AutoMigrate）
├── config/config.go                # 配置结构体定义
├── internal/
│   ├── app/deps.go                 # 进程级依赖（Cfg/DB/RDB/Cache/WxPay）；main 中 Register，业务代码用 Must()
│   ├── initialize/                 # 初始化逻辑
│   │   ├── init.go                 # DB/Redis/Admin/Config/…
│   │   ├── seed.go                 # 展示数据（商品/文章/优惠券/轮播图）
│   │   ├── automigrate.go          # GORM 模型列表 + RunSchemaAutoMigrate
│   │   └── sql_migrate.go          # golang-migrate 嵌入 FS + RunAllSchemaMigrations
│   ├── migratefs/                  # 嵌入的 *.sql（新增 DDL 追加 000002_xxx.up/.down.sql）
│   ├── model/                      # GORM 数据模型（95 张表，以 AutoMigrate / scripts/doc-metrics.sh 为准）
│   ├── service/                    # 业务逻辑层
│   ├── handler/                    # HTTP 处理器（Controller）
│   │   ├── compat/shopxo/*.go      # ShopXO uni-app /api.php（routeMap 登记 s= 动作，调度见 compat.go）
│   │   └── ...
│   ├── router/router.go            # 路由注册（353 条 Gin；全站合计 395 见 HANDOVER.md / doc-metrics.sh）
│   └── middleware/                  # 中间件
│       ├── jwt.go                  # 用户 JWT 认证
│       ├── admin_auth.go           # 管理员 JWT 认证
│       ├── cors.go                 # 跨域
│       └── logger.go               # 请求日志
├── pkg/                            # 公共包
│   ├── auth/jwt.go                 # JWT 生成/解析
│   ├── response/response.go        # 统一响应格式
│   └── wechat/client.go            # 微信支付客户端
├── web/                            # PC 前台（Next.js）
├── admin/                          # 管理后台（Next.js + Ant Design）
└── static/                         # DIY/Form 构建产物
```

## 添加新接口的流程

以"添加商品标签功能"为例：

### 1. 定义 Model

```go
// internal/model/tag.go
package model

type GoodsTag struct {
    ID      uint   `json:"id" gorm:"primaryKey"`
    GoodsID uint   `json:"goods_id" gorm:"index"`
    Name    string `json:"name" gorm:"size:32;not null"`
}
```

### 2. 注册 AutoMigrate

在 `internal/initialize/automigrate.go` 的 `autoMigrateModelList()` 中加入 `&model.GoodsTag{}`；**若用 SQL 独占演进**，同步在 `internal/migratefs` 增加版本文件，并考虑 `GOSHOP_DISABLE_AUTOMIGRATE=true`。

### 3. 编写 Service

```go
// internal/service/tag.go
package service

import (
    "github.com/zhangpanda/goshop/internal/app"
    "github.com/zhangpanda/goshop/internal/model"
)

func GetGoodsTags(goodsID uint) ([]model.GoodsTag, error) {
    var tags []model.GoodsTag
    return tags, app.Must().DB.Where("goods_id = ?", goodsID).Find(&tags).Error
}

func AddGoodsTag(goodsID uint, name string) error {
    return app.Must().DB.Create(&model.GoodsTag{GoodsID: goodsID, Name: name}).Error
}
```

### 4. 编写 Handler

```go
// internal/handler/tag.go
package handler

func GetGoodsTagsHandler(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
    tags, err := service.GetGoodsTags(uint(id))
    if err != nil { response.Fail(c, 500, err.Error()); return }
    response.OK(c, tags)
}
```

### 5. 注册路由

在 `internal/router/router.go` 中添加：

```go
api.GET("/goods/:id/tags", handler.GetGoodsTagsHandler)
```

### 6. 前端调用

```typescript
const tags = await api.get<Tag[]>(`/goods/${id}/tags`)
```

## 添加新的管理后台页面

管理后台在 `admin/` 目录下，使用 Next.js App Router + Ant Design。

### 1. 创建页面文件

```
admin/src/app/(dashboard)/tags/page.tsx
```

### 2. 使用 CrudPage 组件

大部分管理页面使用统一的 `CrudPage` 组件，只需定义列和表单：

```tsx
import CrudPage from '@/components/CrudPage'

export default function TagsPage() {
  return <CrudPage
    title="标签管理"
    apiBase="/api/admin/tags"
    columns={[
      { title: 'ID', dataIndex: 'id', width: 80 },
      { title: '名称', dataIndex: 'name' },
    ]}
    formFields={[
      { name: 'name', label: '标签名称', required: true },
    ]}
  />
}
```

### 3. 添加菜单

在 `admin/src/components/AdminShell.tsx` 的菜单配置中添加新项。

## 添加新的支付方式

1. 在 `pkg/` 下创建支付驱动包
2. 在 `internal/service/pay.go` 中注册驱动
3. 在管理后台「支付方式」中配置

## 配置说明

`config.yaml` 分组：

| 分组 | 说明 | 关键配置 |
|------|------|----------|
| server | 服务 | port, mode(debug/release) |
| db | MySQL | host, port, user, password, dbname |
| redis | Redis | host, port, password |
| jwt | 认证 | secret, expire(小时) |
| wechat | 微信支付 | app_id, mch_id, mch_api_key, private_key, notify_url |

## 数据库配置（后台可改）

通过管理后台「系统配置」修改，存储在 `configs` 表中：

| 分组 | 说明 |
|------|------|
| base | 站点名称/Logo/公告/页脚 |
| site | 站点状态/注册方式/登录方式 |
| seo | SEO 标题/关键词/描述 |
| order | 自动关闭时间/自动确认收货天数 |
| email | SMTP 配置 |
| sms | 短信平台配置 |
| app | 搜索开关/客服电话/邮箱/工作时间 |

## JWT 认证机制

- 用户 token：`IsAdmin=false`，通过 `JWTAuth()` 中间件验证
- 管理员 token：`IsAdmin=true`，通过 `AdminAuth()` 中间件验证
- 两种 token 互不通用，防止越权
- 默认过期时间 72 小时，在 `config.yaml` 的 `jwt.expire` 中配置

## ShopXO 兼容层（`/api.php`）

> **产品目标**是以 **ShopXO v6.8.0 PHP 部署** 为基线实现 **Go 替换**：`/api.php` 保留同名入口与 `s=` 路由习惯，便于 shopxo-uniapp 等前端**仅换站点根地址**即可对接。GoShop 与 ShopXO 无代码隶属关系，行为以本仓库实现与测试为准；**不在替换承诺内的能力**（插件市场、PHP 插件运行时等）见根目录 `README.md` / `HANDOVER.md`。

`/api.php?s=controller/action` 入口由 `internal/compat/shopxo` 注册，将 ShopXO 风格的请求映射为 GoShop 内部调用。

如需添加新的兼容接口，在 `routeMap` 中注册即可：

```go
var routeMap = map[string]gin.HandlerFunc{
    "your/action": yourHandler,
}
```

需要登录的接口在 `authRequiredRoutes` 中标记。

**与官方约定对齐（可复现）**：

1. **uni-app 静态路由 ⊆ Go `routeMap`**：CI 任务 **`shopxo-parity`** 会 shallow clone [shopxo-uniapp](https://github.com/gongfuxiang/shopxo-uniapp) 并执行 `python3 scripts/extract_shopxo_uniapp_routes.py /tmp/shopxo-uniapp --fail-on-missing`（**不含** `plugins/index` 类插件调用）。
2. **冻结 87 条 ⊆ PHP `app/api/controller` 且 ⊆ Go**：同一任务再 clone [ShopXO](https://github.com/gongfuxiang/shopxo) 的 **`v6.8.0`**，执行 `python3 scripts/extract_shopxo_php_api_routes.py /tmp/shopxo-php --fail-uniapp-contract`，确保 `scripts/data/shopxo_uniapp_normal_routes.txt` 中每条在 **PHP 扫描结果与 `compat.go` routeMap** 中均存在（方法名映射：`LoginVerifySend` → `loginverifysend`）。
3. **全量 HTTP 冒烟**：`scripts/integration_test.sh` 末尾执行 `scripts/smoke_shopxo_s_routes.py`（单独跑：`BASE=http://127.0.0.1:8080 python3 scripts/smoke_shopxo_s_routes.py`）。
4. **与 PHP 的 JSON 形状对拍（本地/双活环境）**：起好 PHP 与 Go 两套站点后，`SHOPXO_PHP_BASE=… SHOPXO_GO_BASE=… python3 scripts/shopxo_dual_json_diff.py [--fail-on-shape]`，按 `scripts/data/shopxo_json_diff_samples.json` 对比 `data` 的键路径（不比较数值）；扩充样本即可加大覆盖面。
