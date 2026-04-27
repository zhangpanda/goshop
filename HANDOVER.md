# GoShop 项目交接文档

## 项目概述
用 Go 实现的电商系统，开发时对照 ShopXO v6.8.0（PHP）的商家主路径与数据模型，前后端分离；已开源发布。
功能覆盖率 **约 97%** 指相对 ShopXO **商家主路径**的人工核对结论（不含下文「产品边界」中刻意不对齐的能力；分级见 `docs/shopxo-admin-parity.md`，**非**第三方审计）。

**ShopXO** 名称仅用于说明兼容与差异，不代表官方合作或商标授权。

## 产品边界（刻意不对齐 ShopXO）
- **插件在线市场 / ShopXO 式应用商店**：**不计划**作为内置系统能力实现；管理端相关 Tab 仅为布局兼容或占位。**是否在未来单独跟进 ShopXO 生态变化**，由后续版本与产品决策另定，**不承诺**与官方商店同步。
- **PHP 包管理 / 在线升级 / 运行任意 PHP 插件字节码**：不在产品范围内；发版与扩展走 Go 工程自有方式（API、配置、自建模块等）。

## 仓库
- GitHub: `git@github.com:zhangpanda/goshop.git`
- Gitee: `git@gitee.com:rilegouasas/goshop.git`
- 对照 ShopXO 源码时请自行准备官方发行版（如 v6.8.x），本仓库不随附 ShopXO 代码。

## 技术栈
- 后端: Go 1.25 + Gin + GORM + MySQL + Redis(可选)（以 `go.mod` 为准）
- 管理后台: Next.js 15 + Ant Design 5（端口3010）
- PC前台: Next.js 15 + Tailwind 4 + framer-motion（端口3000）
- 手机端: 可选 shopxo-uniapp 等 + 后端 `/api.php` 兼容层（见 `docs/uniapp-guide.md`）

## 启动
```bash
./bin/goshop                    # 后端 :8080（首次自动建表+seed）
cd admin && npm run dev         # 管理后台 :3010（admin/admin123）
cd web && npm run dev           # PC前台 :3000
```

## 当前状态（v1.5.3, 2026-04-26）

### 核心数据
| 指标 | 数值 |
|------|------|
| Go 后端代码 | 17285 行（`internal`+`pkg`+`cmd`+`config`+`global`，不含 `*_test.go`） |
| Gin HTTP 注册 | **392**（`internal/router/router.go` **350** + `diyapi_compat` **41** + `/api.php` **1**） |
| ShopXO `api.php` | **82** 个 `s=` 动作（`routeMap`，单入口 `Any`） |
| 数据库表 | **95**（`cmd/server/main.go` 中 `AutoMigrate` 的去重模型数） |
| 管理后台页面 | **70**（`admin/src/app/**/page.tsx`） |
| 管理后台组件 | **12**（`admin/src/components/*` 顶层文件） |
| PC前台页面 | **24**（`web/src/app/**/page.tsx`） |
| Go 单元测试 | **47**（`^func Test`，全仓 `*_test.go`） |
| 集成/验收脚本 | **3**（`scripts/integration_test.sh`、`sandbox_pay_test.sh`、`distribution_test.sh`） |

> 上表为对外文档的**权威口径**。更新实现后请跑 **`scripts/doc-metrics.sh`** 刷新数字，并同步 `README.md` / `docs/*`，避免漂移。

### 本轮完成的全部工作

#### Redis 可选
- `pkg/cache` 抽象层：Cache 接口 + RedisCache + MemoryCache（带 Close/stop）
- 未配置 Redis 自动 fallback 内存缓存

#### 支付系统
- 支付宝：5种场景驱动 + RSA2 验签回调 + json.Marshal 防注入
- 支付沙盒：`payment.sandbox: true` 开启，包装真实驱动走完签名逻辑，只跳过 HTTP
- 沙盒回调 `/api/pay/sandbox/callback` 模拟第三方通知

#### 营销功能
- 秒杀：gorm.Expr 原子扣库存 + 每人限购
- 拼团：开团/参团/自动成团，gorm.Expr 原子操作
- 分销：二级分销（佣金比例可配置），提现申请/审核，订单完成自动结算

#### RBAC 权限（从样子货变为真实生效）
- admin 路由拆为 14 个权限组，每组挂载 AdminPower 中间件
- 超级管理员 `"*"` 自动通过所有检查

#### 订单拆分
- CreateOrder 先调 SplitOrderByWarehouse，多仓库自动拆单

#### 短信 & 邮件（从空壳变为真实发送）
- 短信：阿里云 SMS API（HMAC-SHA1 签名），未配置时优雅降级
- 邮件：SMTP 发送（支持 465 SSL + 25/587 STARTTLS），邮件测试接口
- 验证码：自动识别邮箱/手机号，走对应通道

#### 验证码加固
- 4位 → 6位（100万种组合）
- IP 限速：同 IP 60秒内最多 5 次

#### 管理后台 ID 解析（从裸数字变为显示名称）
- 通用 hook：`useUserMap` + `useGoodsMap`（批量请求，50ms 合并，全局缓存）
- 后端：`AdminGetUsers` 和 `GetGoodsList` 支持 `?ids=1,2,3` 批量查询
- 修复 13 个页面：订单/售后/评价/浏览/收藏/购物车/消息/搜索记录/积分日志/支付日志/表单数据/用户地址/仓库商品
- 分销管理页：添加分销商弹窗（搜索用户）、提现审核、佣金比例设置

#### Bug 修复
- 竞态：秒杀 sold / 拼团 joinCount / 钱包余额 → gorm.Expr
- JSON 注入：支付宝 biz_content + 快递100 param → json.Marshal
- nil 指针：微信 PayNotify
- 参数校验：handler 中 ParseUint 错误处理
- 错误处理：AutoMigrate / sqlDB 错误检查

#### CI
- GitHub Actions 加 MySQL 8.0 容器跑集成测试
- `admin-e2e` Job：`GOSHOP_E2E=1` 启动后端 + Playwright 跑管理端营销等用例（见 `admin/e2e/`）
- Redis 不需要（内存缓存 fallback）
- **Fmt**：CI 执行 `gofmt -s -l .`，提交前本地可跑 `gofmt -s -w .` 避免失败
- **集成 Job 内 `config.yaml`**：增加 `payment.sandbox: true`，并以 **`GOSHOP_PAYMENT_SANDBOX=1`** 调用 `scripts/integration_test.sh`，覆盖 **ShopXO 多订单 + 沙盒回调** 路径

#### 文档
- docs/shopxo-admin-parity.md — 管理端与 ShopXO 后台模块对齐清单（可核对）
- docs/migration-from-shopxo.md — ShopXO v6.x 迁移指南（SQL 与 `config/shopxo.sql` 表结构对齐，含 SKU / 订单明细 / 订单地址 JSON）
- docs/api.md — 补充秒杀/拼团/分销/统一支付等新 API
- docs/deployment.md — Redis 可选 + 支付宝配置 + **Nginx `X-Forwarded-Proto`** + **关闭 AutoMigrate 时 `orders.payment_id` 的 ALTER 示例**
- README.md 全面更新
- scripts/MANUAL_VERIFY_PROXY.md — **反代/HTTPS 下 ShopXO 支付、回调、DIY 附件与 DB 的手工验收清单**

#### 支付与 ShopXO / DIY 兼容补充（同日后端提交）
- **订单 `payment_id`**：`orders` 表增加字段，记录用户选用的支付方式主键；`UnifiedPayReq` 增加 `payment_id`（JSON）映射 `PaymentRecordID`，发起支付时回写订单；`PayLogSuccess` 在合并支付成功时亦回写子订单的 `payment_id`。
- **合并支付**：`MultiOrderUnifiedPay`（`internal/service/pay_log.go`）支持多笔待付订单：线下/钱包批量更新；微信支付宝等先 `CreatePayLog`，第三方 `out_trade_no` 使用 **`PayLog.pay_no`**，金额为日志汇总实付。
- **支付回调**：`HandlePayNotify` 先按 **`order_no`** 匹配单笔订单（与旧链路一致）；未命中则按 **`pay_no`** 调用 `PayLogSuccess`（合并支付）。
- **ShopXO uni-app**：`order/pay` 支持多个 `ids`（或 `id` 逗号分隔），单笔仍走 `UnifiedPay`，多笔走 `MultiOrderUnifiedPay`。订单列表/详情中 `payment_id`、`payment_name` 优先取订单上的 `payment_id`，为 0 时回退 `DefaultPaymentIDForShopXO()`。
- **DIY 管理端 URL**：`baseURL()` 识别反向代理 **`X-Forwarded-Proto`** / **`X-Forwarded-Host`**（取逗号分隔首段），避免 TLS 终止后仍生成 `http://` 错误链接。
- **附件远程抓取**：`POST /api/attachmentapi/catch` 实现受限 HTTP 拉取（仅 http/https、拒绝常见内网解析结果、超时 20s、体 ≤5MB、白名单图片扩展名或 `Content-Type`），落盘至 `uploads/YYYY/MM/DD/` 并写入 **`Attachment`** 表；失败 URL 跳过，成功项放入响应 `data`。
- **部署**：重启后端后 GORM `AutoMigrate` 会为 `orders` 增加 `payment_id` 列；若禁用自动迁移需自行 `ALTER TABLE` 对齐模型（见 `docs/deployment.md`）。

#### CI 与集成测试补充（v1.5.2）
- **`EnsureDefaultPayments()`**（`internal/initialize/seed.go`，在 `main` 中 `InitDefaultSeedData` 之后调用）：当库中**没有任何** `payments` 记录时，自动插入 **12 条**默认渠道（线下、钱包、微信 JSAPI/H5/APP/扫码、支付宝 H5/PC/APP/小程序、当面付、PayPal），与 `payment_driver.go` 注册名一一对应；**已有数据的库不会自动补行**，需后台手工新增或清表后重建。
- **`scripts/integration_test.sh`**：校验 **`GET /api/payments`** 含上述 `payment_key`；对首单走 **`POST /api/pay/unified` 线下支付**；另建一单测取消。末尾仍有 **ShopXO `order/pay` 多订单线下**；若 **`GOSHOP_PAYMENT_SANDBOX=1`** 且 `payment.sandbox=true`，再跑多订单微信沙盒回调。
- **手工回归**：生产或预发在反向代理 + HTTPS 下按 **`scripts/MANUAL_VERIFY_PROXY.md`** 逐项验收（与自动化互补）。

#### 管理后台纠偏（v1.5.3）
- **批量导出**：`ExportData`（`internal/service/extend.go`）支持请求体 **`ids`**，仅导出勾选行；**`BatchActions`** 改为 `fetch` 下载 CSV，增加 **`exportType`**（`orders` / `users` / `goods`），与 **`ExportButton`** 行为一致。订单 / 用户 / 商品列表已接入「导出选中」。
- **用户批量操作**：新增 **`DELETE /admin/users/:id`**，实现为 **`AdminDisableUser`**（`status=0`，不物理删除，保留订单关联）；列表上按钮文案为 **批量禁用**，避免与真实删库混淆。
- **售后列表**：移除无实际操作的 **`BatchActions`** 与行多选，避免「已选 N 条」无按钮。
- **语言与货币**：管理端 **`GET/POST /admin/multilingual`**、**`GET/POST /admin/currency`**；前台菜单 **系统 → 语言与货币**（`admin/src/app/(dashboard)/locale/page.tsx`）。**`GetCurrencyConfig`** 从配置读取 **`currency_rate`**。
- **角色与插件**：**`GET /admin/roles/:id/plugins`** + **`GetRolePluginIDs`**；RBAC 角色表增加 **「分配插件」**，对应 **`PUT /admin/roles/:id/plugins`**。**应用商店 Tab** 仍为占位（见「产品边界」，无内置市场计划）。

## 待办

### P3 - 未来
1. **多语言前端** — i18n 国际化
2. **性能优化** — 数据库索引、热点数据缓存
3. **优雅关停** — 信号处理 + 请求排空
4. **PayPal 对接** — 需海外商户账号

## 关键文件

### 后端
```
cmd/server/main.go                 # 入口（含 EnsureDefaultPayments）
internal/initialize/seed.go        # 商品等 seed + EnsureDefaultPayments
pkg/cache/cache.go                 # 缓存抽象层(Redis/Memory)
internal/router/router.go          # 路由(14个RBAC权限组)
internal/service/payment_driver.go # 12种支付驱动 + 沙盒
internal/service/distribution.go   # 分销(佣金结算/提现)
internal/service/seckill.go        # 秒杀
internal/service/group_buy.go      # 拼团
internal/service/order.go          # 订单(含拆单)
internal/service/pay_log.go        # PayLog、合并支付 MultiOrderUnifiedPay、PayLogSuccess
internal/service/notify.go         # 短信(阿里云) + 邮件(SMTP)
internal/service/logistics.go      # 物流轨迹(快递100)
internal/middleware/admin_auth.go  # AdminAuth + AdminPower
internal/handler/diyapi_compat.go # diyapi + attachmentapi（含 baseURL、attachmentApiCatch）
internal/handler/shopxo_compat.go  # ShopXO 兼容（含多订单 order/pay）
internal/service/extend.go         # 导出 CSV（含 ids）、多语言/货币配置
internal/service/user.go         # AdminDisableUser
```

### 前端
```
admin/src/app/(dashboard)/locale/page.tsx  # 语言与货币
admin/src/components/BatchActions.tsx      # 批量删除/禁用/导出选中
admin/src/lib/useIdMap.ts             # 通用ID→名称解析hook
admin/src/components/ParamsEditor.tsx  # 商品参数结构化编辑器
admin/src/components/JsonConfigEditor.tsx # JSON可视化配置
admin/src/app/(dashboard)/distribution/page.tsx # 分销管理(3个Tab)
```

### 测试
```
scripts/integration_test.sh      # 核心 API + ShopXO 多单线下；可选 GOSHOP_PAYMENT_SANDBOX=1 跑多单沙盒回调
scripts/MANUAL_VERIFY_PROXY.md   # 反代/HTTPS 下手工验收清单（非脚本）
scripts/sandbox_pay_test.sh      # 全渠道沙盒轮询（含 PayPal/当面付）+ 钱包边界
scripts/distribution_test.sh     # 分销完整链路测试
.github/workflows/ci.yml       # gofmt / vet / test / 集成（含 payment.sandbox）
```
