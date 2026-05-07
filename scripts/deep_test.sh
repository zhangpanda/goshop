#!/bin/bash
# 本地深度校验：静态检查 + 全量单元测试（不依赖 MySQL/Redis）。
# 集成/沙盒仍需服务：scripts/integration_test.sh、scripts/sandbox_pay_test.sh
#
# 用法: bash scripts/deep_test.sh
# 可选: GOSHOP_TEST_RACE=1 — 追加 -race（较慢，约数分钟）

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# 排除误扫入的 web/node_modules 下 Go 包（逐包传参，避免某些 shell 把换行包成单个参数）
pkgs=()
while IFS= read -r line; do
  [[ -n "$line" ]] && pkgs+=("$line")
done < <(go list ./... | grep -v '/node_modules/' || true)

echo "========== go vet =========="
go vet "${pkgs[@]}"

echo ""
echo "========== go test (全模块) =========="
go test "${pkgs[@]}" -count=1

if [ "${GOSHOP_TEST_RACE:-}" = "1" ]; then
  echo ""
  echo "========== go test -race =========="
  go test "${pkgs[@]}" -count=1 -race -timeout 5m
fi

echo ""
echo "========== 完成 =========="
echo "若已启动 API，可另执行: BASE=http://localhost:8080 bash scripts/integration_test.sh"
