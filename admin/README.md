# GoShop 管理后台

基于 Next.js 15 + Ant Design 5 的管理后台，68 个页面，100% 对齐 ShopXO 后台功能。

## 开发

```bash
npm install
npm run dev    # http://localhost:3001
```

默认管理员：admin / admin123

## 构建

```bash
npm run build
npm start
```

## 目录结构

```
src/
├── app/
│   ├── (dashboard)/    # 管理页面（68个）
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
