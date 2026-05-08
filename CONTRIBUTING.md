# 贡献指南

感谢你对 GoShop 的关注！欢迎提交 Issue 和 Pull Request。

## 开发环境

- Go **1.25.10**（`go.mod` 的 `go` + **`toolchain`**；CI **`setup-go`**；**`Dockerfile`** → `golang:1.25.10-alpine`；可选 **`mise.toml`** / **`.tool-versions`**）
- Node.js 20+
- MySQL 8.0+（5.7 亦通常可用）
- Redis 6+（可选；不配置则使用内存缓存）

升级 Go 补丁时请按 **`docs/development.md`** 或 **`python3 scripts/sync_go_toolchain.py --write`** **一次性回写**（CI 会跑 `scripts/sync_go_toolchain.py` **校验**），勿只改 `go.mod`。

## 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/zhangpanda/goshop.git
cd goshop

# 2. 配置
cp config.yaml.example config.yaml
# 编辑 config.yaml 填写数据库和 Redis 连接信息
# **勿将 config.yaml、证书、真实密钥提交到 git**（仓库已 .gitignore `config.yaml` 与 `cert/`）

# 3. 启动后端
go run ./cmd/server/

# 4. 启动 Web 前端
cd web && npm install && npm run dev

# 5. 启动管理后台
cd admin && npm install && npm run dev
```

管理后台 E2E（Playwright）：另开终端在仓库根目录 `GOSHOP_E2E=1 go run ./cmd/server/main.go`，再 `cd admin && npx playwright install chromium && npm run test:e2e`。详见 [admin/README.md](admin/README.md)。

Go 后端单测：**日常** `bash scripts/quick_test.sh`（`go vet` + `go test`，无 race，较快）；**对齐 CI 强度** `bash scripts/ci_test.sh`（再多跑全量 `-race`，较慢）。兼容旧命令：`bash scripts/deep_test.sh`；`GOSHOP_TEST_RACE=1 bash scripts/deep_test.sh` 等同 `ci_test`。

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
