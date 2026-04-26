# 贡献指南

感谢你对 GoShop 的关注！欢迎提交 Issue 和 Pull Request。

## 开发环境

- Go 1.23+
- Node.js 20+
- MySQL 8.0+
- Redis 7+

## 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/zhangpanda/goshop.git
cd goshop

# 2. 配置
cp config.yaml.example config.yaml
# 编辑 config.yaml 填写数据库和 Redis 连接信息

# 3. 启动后端
go run ./cmd/server/

# 4. 启动 Web 前端
cd web && npm install && npm run dev

# 5. 启动管理后台
cd admin && npm install && npm run dev
```

## 提交规范

- feat: 新功能
- fix: 修复 Bug
- docs: 文档更新
- refactor: 代码重构
- test: 测试相关
- chore: 构建/工具变更

示例：`feat: 添加商品规格模板管理`

## Pull Request

1. Fork 本仓库
2. 创建特性分支 `git checkout -b feat/your-feature`
3. 提交代码并推送
4. 创建 Pull Request，描述你的改动

## 代码规范

- Go 代码使用 `gofmt` 格式化
- 前端代码遵循项目已有的 ESLint 配置
- 新增 API 需要在 README 中补充文档
