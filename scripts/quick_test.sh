#!/bin/bash
# 本地快速校验：go vet + 全量单元测试（排除 node_modules 误入包），无 -race。
# 日常改代码跑这个即可（通常比含 race 的快一个数量级）。
#
# 用法: bash scripts/quick_test.sh
# 对齐 CI 的 Go 测 + race: bash scripts/ci_test.sh
# 兼容旧入口: bash scripts/deep_test.sh

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

pkgs=()
while IFS= read -r line; do
  [[ -n "$line" ]] && pkgs+=("$line")
done < <(go list ./... | grep -v '/node_modules/' || true)

echo "========== go vet (quick) =========="
go vet "${pkgs[@]}"

echo ""
echo "========== go test (无 -race) =========="
go test "${pkgs[@]}" -count=1

echo ""
echo "========== quick 完成 =========="
echo "若需 -race（较慢）: bash scripts/ci_test.sh"
echo "若已启动 API: BASE=http://localhost:8080 bash scripts/integration_test.sh"
