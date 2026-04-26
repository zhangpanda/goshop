# GoShop Web 前台

基于 Next.js 15 + Tailwind CSS 的 PC 端商城前台，Apple 风格 UI。

## 页面

| 路由 | 页面 |
|------|------|
| / | 首页（轮播图 + 商品展示） |
| /products | 商品列表（分类筛选/搜索/排序） |
| /products/:id | 商品详情（SKU 选择/加购/收藏） |
| /articles | 文章列表 |
| /articles/:id | 文章详情 |
| /cart | 购物袋 |
| /checkout | 结算 |
| /checkout/pay | 支付 |
| /login | 登录/注册 |
| /account | 用户中心（订单/地址/收藏/积分/消息等） |
| /story | 品牌故事 |
| /support | 支持与帮助 |

## 开发

```bash
npm install
npm run dev    # http://localhost:3000
```

API 请求通过 `next.config.js` 的 rewrites 代理到后端 `http://localhost:8080`。

## 构建

```bash
npm run build
npm start      # 生产模式
```

## 目录结构

```
src/
├── app/           # 页面（App Router）
├── components/    # 公共组件（Header/Footer）
└── lib/
    ├── api.ts         # API 客户端（含 401 自动跳转、formatPrice）
    └── site-config.tsx # 站点配置 Context Provider
```

## 自定义

- 修改颜色：`src/app/globals.css` 中的 CSS 变量
- 修改导航：后台「导航管理」配置，无需改代码
- 修改联系方式：后台「系统配置 → app」配置
