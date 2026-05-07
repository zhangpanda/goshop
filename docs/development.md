# GoShop 二次开发指南

## 项目架构

```
goshop/
├── cmd/server/main.go              # 入口：加载配置 → 初始化DB/Redis → 建表 → Seed → 启动HTTP
├── config/config.go                # 配置结构体定义
├── global/global.go                # 全局变量（DB/Redis/Config/WxPay）
├── internal/
│   ├── initialize/                 # 初始化逻辑
│   │   ├── init.go                 # DB/Redis/Admin/Config/Powers/Navigation
│   │   └── seed.go                 # 展示数据（商品/文章/优惠券/轮播图）
│   ├── model/                      # GORM 数据模型（95 张表，以 AutoMigrate / scripts/doc-metrics.sh 为准）
│   ├── service/                    # 业务逻辑层
│   ├── handler/                    # HTTP 处理器（Controller）
│   │   ├── compat/shopxo/*.go      # ShopXO uni-app 兼容（82 个 s= 动作，调度见 compat.go）
│   │   └── ...
│   ├── router/router.go            # 路由注册（352 条 Gin；全站合计 394 见 HANDOVER.md / doc-metrics.sh）
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

在 `cmd/server/main.go` 的 `AutoMigrate()` 中添加 `&model.GoodsTag{}`。

### 3. 编写 Service

```go
// internal/service/tag.go
package service

import "github.com/zhangpanda/goshop/global"
import "github.com/zhangpanda/goshop/internal/model"

func GetGoodsTags(goodsID uint) ([]model.GoodsTag, error) {
    var tags []model.GoodsTag
    return tags, global.DB.Where("goods_id = ?", goodsID).Find(&tags).Error
}

func AddGoodsTag(goodsID uint, name string) error {
    return global.DB.Create(&model.GoodsTag{GoodsID: goodsID, Name: name}).Error
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

## ShopXO 兼容层

> 此处「兼容」指 **HTTP 形态与移动端对接习惯**，便于衔接 shopxo-uniapp 等前端；GoShop 与 ShopXO 无代码隶属关系，行为以本仓库实现为准。

`/api.php?s=controller/action` 入口由 `internal/compat/shopxo` 注册，将 ShopXO 风格的请求映射为 GoShop 内部调用。

如需添加新的兼容接口，在 `routeMap` 中注册即可：

```go
var routeMap = map[string]gin.HandlerFunc{
    "your/action": yourHandler,
}
```

需要登录的接口在 `authRequiredRoutes` 中标记。
