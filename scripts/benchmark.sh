#!/bin/bash
# GoShop 核心 API 压测脚本
set -e
cd "$(dirname "$0")/.."
HEY=$(which hey 2>/dev/null || echo "$(go env GOPATH)/bin/hey")
BASE=${BASE:-http://127.0.0.1:8080}

# 重启后端（release 模式）
kill $(lsof -ti :8080) 2>/dev/null || true
sleep 1
GIN_MODE=release GOSHOP_E2E=1 nohup ./bin/goshop > /tmp/goshop-bench.log 2>&1 &
sleep 3

# 获取 token
TOKEN=$(curl -s -X POST $BASE/api/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123","captcha_key":"t","captcha_code":"0"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

echo "=== GoShop 压测 (200并发, 2000请求) ==="
echo ""

echo "--- 1. GET /api/categories (公开，无鉴权) ---"
$HEY -n 2000 -c 200 $BASE/api/categories 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "--- 2. GET /api/goods (商品列表) ---"
$HEY -n 2000 -c 200 "$BASE/api/goods?page=1&page_size=20" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "--- 3. GET /api/goods/1 (商品详情) ---"
$HEY -n 2000 -c 200 $BASE/api/goods/1 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "--- 4. POST /api/admin/login (登录) ---"
$HEY -n 1000 -c 100 -m POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123","captcha_key":"t","captcha_code":"0"}' \
  $BASE/api/admin/login 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "--- 5. GET /api/admin/orders (管理端订单列表) ---"
$HEY -n 2000 -c 200 -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/admin/orders?page=1&page_size=20" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "--- 6. GET /api/admin/dashboard (仪表盘统计) ---"
$HEY -n 1000 -c 100 -H "Authorization: Bearer $TOKEN" \
  $BASE/api/admin/dashboard 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "--- 7. GET /api/site-config (站点配置) ---"
$HEY -n 2000 -c 200 $BASE/api/site-config 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
echo ""

echo "=== 压测完成 ==="
