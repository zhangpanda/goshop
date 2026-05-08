#!/usr/bin/env bash
# 统计仓库规模指标，供 HANDOVER.md / README 等文档对齐。无网络依赖。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "=== GoShop doc metrics ($(date -I 2>/dev/null || date '+%Y-%m-%d')) ==="
echo

echo -n "AutoMigrate 模型类型数（≈ 表数）: "
grep -oE '&model\.[A-Za-z0-9_]+' internal/initialize/automigrate.go | sort -u | wc -l | tr -d ' '

echo -n "internal/router/router.go Gin 方法注册数: "
grep -oE '\.(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|Any)\(' internal/router/router.go | wc -l | tr -d ' '

echo -n "diyapi_compat SetupDiyApiCompat 内 g.GET/g.POST 等: "
sed -n '/^func SetupDiyApiCompat/,/^}$/p' internal/compat/shopxo/diyapi.go | grep -E '^\s+g\.(GET|POST|PUT|PATCH|DELETE)\(' | wc -l | tr -d ' '

echo -n "ShopXO routeMap 条目数（s= 动作）: "
sed -n '/^var routeMap = map\[string\]gin.HandlerFunc{/,/^}$/p' internal/compat/shopxo/compat.go | grep -c '^\t"[^"]*":' || true

echo -n "Gin 注册合计（router + diy 组 + /api.php Any×1）: "
R="$(grep -oE '\.(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|Any)\(' internal/router/router.go | wc -l | tr -d ' ')"
D="$(sed -n '/^func SetupDiyApiCompat/,/^}$/p' internal/compat/shopxo/diyapi.go | grep -E '^\s+g\.(GET|POST|PUT|PATCH|DELETE)\(' | wc -l | tr -d ' ')"
echo "$((R + D + 1))"

echo
echo -n "admin/src/app page.tsx: "
find admin/src/app -name 'page.tsx' 2>/dev/null | wc -l | tr -d ' '

echo -n "web/src/app page.tsx: "
find web/src/app -name 'page.tsx' 2>/dev/null | wc -l | tr -d ' '

echo -n "admin/src/components 下文件数: "
find admin/src/components -maxdepth 1 -type f 2>/dev/null | wc -l | tr -d ' '

echo
echo -n "internal/handler *.go（含测试）: "
find internal/handler -maxdepth 1 -name '*.go' | wc -l | tr -d ' '

echo -n "internal/service *.go（含测试）: "
find internal/service -maxdepth 1 -name '*.go' | wc -l | tr -d ' '

echo -n "internal/model *.go: "
find internal/model -maxdepth 1 -name '*.go' | wc -l | tr -d ' '

echo
echo -n "^func Test 个数: "
grep -r --include='*_test.go' -E '^func Test' internal pkg cmd config 2>/dev/null | wc -l | tr -d ' '

echo -n "Go 代码行数（internal+pkg+cmd+config，不含 *_test.go）: "
find internal pkg cmd config -name '*.go' ! -name '*_test.go' 2>/dev/null | while read -r f; do wc -l <"$f"; done | awk '{s+=$1} END{print s}'

echo
echo "集成测试脚本: scripts/integration_test.sh, scripts/sandbox_pay_test.sh, scripts/distribution_test.sh"
