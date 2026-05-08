#!/bin/bash
# 对齐 backend job 的 Go 校验：先 quick（vet + 普通 test），再全量 -race（较慢）。
# 发版前、大改并发相关代码时建议在本地跑一遍。
#
# 用法: bash scripts/ci_test.sh

set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
bash "$HERE/quick_test.sh"

ROOT="$(cd "$HERE/.." && pwd)"
cd "$ROOT"

pkgs=()
while IFS= read -r line; do
  [[ -n "$line" ]] && pkgs+=("$line")
done < <(go list ./... | grep -v '/node_modules/' || true)

echo ""
echo "========== go test -race（全模块，timeout 5m）=========="
go test "${pkgs[@]}" -count=1 -race -timeout 5m

echo ""
echo "========== ci 本地 Go 测完成 =========="
echo "另有 gofmt/govulncheck/集成见 .github/workflows/ci.yml"
