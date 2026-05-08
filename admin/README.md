# GoShop 管理后台

基于 Next.js 15 + Ant Design 5 的管理后台，**70** 个页面（`src/app/**/page.tsx`）；与 ShopXO 后台为分级对照（见仓库根目录 `docs/shopxo-admin-parity.md`、`HANDOVER.md`；规模统计可跑 `scripts/doc-metrics.sh`）。

## 开发

```bash
npm install
npm run dev    # http://localhost:3010（见 package.json）
```

默认管理员：admin / admin123

## E2E（Playwright）

依赖 **Go 后端**（默认 `http://localhost:8080`，与 `next.config.js` 反代一致；后端工具链为 **Go 1.25.10**，见仓库根 `go.mod` / `Dockerfile`）。自动化下需 **`GOSHOP_E2E=1`** 跳过登录验证码（仅用于本地/CI，**生产勿开**）。

```bash
# 终端 1：仓库根目录，配置好 config.yaml 后
GOSHOP_E2E=1 ./bin/goshop   # 或 go run ./cmd/server/main.go

# 终端 2
cd admin
npm install
npx playwright install chromium   # 首次
npm run test:e2e                  # 会自动起 next dev :3010（若未占用）
# npm run test:e2e:ui            # 调试
```

可选环境变量：`E2E_ADMIN_USER`、`E2E_ADMIN_PASS`（默认 admin / admin123）。

## 构建

```bash
npm run build
npm start
```

## 目录结构

```
src/
├── app/
│   ├── (dashboard)/    # 管理页面（70个）
│   │   ├── layout.tsx  # 侧边栏布局
│   │   ├── page.tsx    # 仪表盘
│   │   ├── goods/      # 商品管理
│   │   ├── orders/     # 订单管理
│   │   └── ...
│   └── login/          # 登录页
├── components/
│   ├── AdminShell.tsx  # 管理后台外壳（侧边栏+顶栏）
│   ├── CrudPage.tsx    # 通用增删改查页面组件
│   ├── RichEditor.tsx  # 富文本编辑器
│   ├── ImageUpload.tsx # 图片上传组件
│   └── ...
└── lib/
    └── admin-auth.tsx  # 管理员认证 Context
```

## 添加新页面

大部分页面使用 `CrudPage` 组件，只需定义列配置和表单字段即可自动生成完整的增删改查页面。详见 [二次开发指南](../docs/development.md)。
