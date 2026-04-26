# GoShop

用 Go 重写的开源电商系统，100% 对齐 ShopXO v6.8.0 功能，前后端分离架构。

> Go 后端代码为独立编写。DIY装修器和Form表单设计器复用 ShopXO 官方子项目（[shopxo-diy](https://github.com/gongfuxiang/shopxo-diy) / [shopxo-form](https://github.com/gongfuxiang/shopxo-form)，MIT），uni-app 移动端可直接对接（[shopxo-uniapp](https://github.com/gongfuxiang/shopxo-uniapp)）。

## 特性

- **Go 后端**：Gin + GORM + MySQL + Redis，308个API路由，12种支付驱动
- **管理后台**：Next.js + Ant Design，68个页面，100%对齐ShopXO后台
- **PC前台**：Next.js + Tailwind CSS，Apple风格UI
- **手机端**：直接复用ShopXO uni-app，兼容层56/56接口通过
- **DIY装修**：集成shopxo-diy可视化拖拽编辑器
- **Form设计**：集成shopxo-form可视化表单设计器

## 截图预览

| 首页 | 商品列表 |
|------|----------|
| ![首页](docs/screenshots/home.png) | ![商品列表](docs/screenshots/products.png) |

| 商品详情 | 管理后台 |
|----------|----------|
| ![商品详情](docs/screenshots/product-detail.png) | ![管理后台](docs/screenshots/admin-login.png) |

## 快速开始

### Docker 一键部署

```bash
git clone https://github.com/zhangpanda/goshop.git
cd goshop
cp config.yaml.example config.yaml
# 编辑 config.yaml，将 db.host 改为 mysql，redis.host 改为 redis
docker compose up -d
# 访问 http://localhost:8080
```

### 本地开发

#### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 5.7+
- Redis 6+

### 1. 启动后端
```bash
git clone https://github.com/zhangpanda/goshop.git
cd goshop
cp config.yaml.example config.yaml  # 修改数据库配置
go build -o bin/goshop cmd/server/main.go
./bin/goshop
# 服务运行在 http://localhost:8080
# 首次启动自动建表、创建默认管理员(admin/admin123)、初始化配置
```

### 2. 启动管理后台
```bash
cd admin
npm install
npm run dev
# 访问 http://localhost:3010/login
# 默认账号: admin / admin123
```

### 3. 启动PC前台
```bash
cd web
npm install
npm run dev
# 访问 http://localhost:3000
```

### 4. 对接uni-app手机端
```bash
# 克隆ShopXO官方uni-app项目
git clone https://github.com/gongfuxiang/shopxo-uniapp.git
# 修改 common/config.js 中的 request_url 为 http://你的IP:8080
# 用HBuilderX打开运行即可
```

### 5. DIY装修器（可选）
```bash
git clone https://github.com/gongfuxiang/shopxo-diy.git
cd shopxo-diy
npm install --legacy-peer-deps
npx vite build
# 将 dist/static/diy 复制到 goshop/static/diy
# 将 dist/index.html 复制到 goshop/static/diy.html
```

### 6. Form表单设计器（可选）
```bash
git clone https://github.com/gongfuxiang/shopxo-form.git
cd shopxo-form
npm install --legacy-peer-deps
npx vite build
# 将 dist/static/form_input 复制到 goshop/static/form_input
# 将 dist/index.html 复制到 goshop/static/form.html
```

## 项目结构

```
goshop/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── handler/                # HTTP处理器 (34文件)
│   ├── service/                # 业务逻辑 (48文件)
│   ├── model/                  # 数据模型 (32文件, 90张表)
│   ├── router/router.go        # 路由注册 (308路由)
│   └── middleware/             # 中间件 (JWT/CORS/操作日志)
├── admin/                      # 管理后台 Next.js (68页面)
├── web/                        # PC前台 Next.js (12页面)
├── static/                     # DIY/Form构建产物
├── pkg/                        # 公共包 (JWT/微信支付/响应)
├── config.yaml                 # 配置文件
└── go.mod
```

## 管理后台功能

| 模块 | 功能 |
|------|------|
| 仪表盘 | 10个报表图表，今日/昨日/周/月切换，待处理事项 |
| 商品 | 商品管理/分类/评论/浏览/收藏/购物车/参数模板/规格模板 |
| 订单 | 订单管理/售后，独立详情页，发货/物流轨迹/打印 |
| 用户 | 用户列表/地址，编辑积分/余额 |
| 网站 | 导航/轮播/链接/快递/地区/筛选价格/自定义页面/支付方式/附件/表单/主题/设计/布局/快捷导航/菜单 |
| 品牌 | 品牌管理/品牌分类 |
| 数据 | 消息/支付日志/积分日志/退款日志/短信日志/邮件日志/错误日志/搜索记录/操作日志 |
| 文章 | 文章管理/文章分类 |
| 手机 | 首页导航/基础配置/小程序配置/用户中心导航/DIY装修 |
| 应用 | 插件安装/卸载/配置/应用商店 |
| 仓库 | 仓库管理/仓库商品 |
| 系统 | 系统配置(12分组77项)/商店信息 |
| 站点 | 站点设置/短信/SEO/邮箱/协议 |
| 权限 | 管理员/角色/权限分配 |
| 工具 | 缓存管理/SQL控制台 |

## 支付驱动

微信支付(JSAPI/H5/APP/Native) + 支付宝(PC/H5/APP/小程序/面对面) + PayPal + 钱包余额 + 线下支付

## 数据库

- 90张表，全部有中文表注释和字段注释
- 首次启动自动建表(AutoMigrate)
- 自动初始化：默认管理员、系统配置(77项)、快递公司(14家)、省市区(3447条)

## 与ShopXO的对比

| 维度 | ShopXO | GoShop |
|------|--------|--------|
| 语言 | PHP (ThinkPHP) | Go (Gin + GORM) |
| 架构 | 前后端一体 | 前后端分离 |
| 后端代码 | ~14.5万行 | ~1.3万行 |
| 管理后台 | PHP模板渲染 | React + Ant Design |
| PC前台 | PHP模板渲染 | Next.js + Tailwind |
| 性能 | 一般 | 高（Go原生并发） |
| 部署 | PHP+Nginx+MySQL | 单二进制+MySQL+Redis |
| 功能覆盖 | 100% | 100% |

## 贡献

欢迎提交 Issue 和 Pull Request，详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 文档

| 文档 | 说明 |
|------|------|
| [API 文档](docs/api.md) | 完整接口文档（公共/用户/管理后台/ShopXO 兼容） |
| [二次开发指南](docs/development.md) | 项目架构、添加接口/页面/支付方式的流程 |
| [部署文档](docs/deployment.md) | Docker/手动/Systemd 部署 + Nginx 配置 |
| [uni-app 对接指南](docs/uniapp-guide.md) | ShopXO uni-app 前端对接说明 |
| [Web 前台](web/README.md) | PC 前台开发说明 |
| [管理后台](admin/README.md) | 管理后台开发说明 |

## 开源协议

[Apache License 2.0](LICENSE)
