#!/bin/bash
# GoShop 核心流程集成测试
# 需要后端服务运行在 localhost:8080
# 用法: ./scripts/integration_test.sh

set -e
BASE="http://localhost:8080"
PASS=0
FAIL=0

assert_code() {
  local name="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "  ✅ $name"
    PASS=$((PASS+1))
  else
    echo "  ❌ $name (expected=$expected, got=$actual)"
    FAIL=$((FAIL+1))
  fi
}

api() { curl -s "$@"; }
code() { echo "$1" | python3 -c "import sys,json;print(json.load(sys.stdin)['code'])"; }

echo "========== GoShop 集成测试 =========="

echo ""
echo "--- 公共接口 ---"
for path in /api/site-config /api/categories /api/goods /api/slides /api/articles /api/article-categories /api/coupons /api/promotions /api/navigations; do
  R=$(api "$BASE$path")
  assert_code "GET $path" "0" "$(code "$R")"
done

echo ""
echo "--- 商品详情 ---"
R=$(api "$BASE/api/goods/1")
assert_code "GET /api/goods/1" "0" "$(code "$R")"

echo ""
echo "--- 注册 ---"
TS=$(date +%s)
R=$(api -X POST "$BASE/api/register" -H 'Content-Type: application/json' -d "{\"username\":\"test${TS}\",\"password\":\"test123456\"}")
assert_code "POST /api/register" "0" "$(code "$R")"

echo ""
echo "--- 登录 ---"
R=$(api -X POST "$BASE/api/login" -H 'Content-Type: application/json' -d "{\"username\":\"test${TS}\",\"password\":\"test123456\"}")
assert_code "POST /api/login" "0" "$(code "$R")"
TOKEN=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")
AUTH="Authorization: Bearer $TOKEN"

echo ""
echo "--- 加购物车 ---"
R=$(api -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' -d '{"goods_id":1,"sku_id":1,"quantity":1}')
assert_code "POST /api/cart" "0" "$(code "$R")"

echo ""
echo "--- 查看购物车 ---"
R=$(api "$BASE/api/cart" -H "$AUTH")
assert_code "GET /api/cart" "0" "$(code "$R")"
CART_ID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data'][0]['id'])")

echo ""
echo "--- 添加地址 ---"
R=$(api -X POST "$BASE/api/address" -H "$AUTH" -H 'Content-Type: application/json' -d '{"name":"测试","phone":"13800000000","province":"北京市","city":"北京市","district":"朝阳区","detail":"测试地址","is_default":true}')
assert_code "POST /api/address" "0" "$(code "$R")"

echo ""
echo "--- 创建订单 ---"
R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":1,\"cart_ids\":[$CART_ID]}")
assert_code "POST /api/orders" "0" "$(code "$R")"
ORDER_ID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")

echo ""
echo "--- 订单列表 ---"
R=$(api "$BASE/api/orders" -H "$AUTH")
assert_code "GET /api/orders" "0" "$(code "$R")"

echo ""
echo "--- 订单详情 ---"
R=$(api "$BASE/api/orders/$ORDER_ID" -H "$AUTH")
assert_code "GET /api/orders/:id" "0" "$(code "$R")"

echo ""
echo "--- 取消订单 ---"
R=$(api -X PUT "$BASE/api/orders/$ORDER_ID/cancel" -H "$AUTH")
assert_code "PUT /api/orders/:id/cancel" "0" "$(code "$R")"

echo ""
echo "--- 收藏商品 ---"
R=$(api -X POST "$BASE/api/favorites/1" -H "$AUTH")
assert_code "POST /api/favorites/:id" "0" "$(code "$R")"

echo ""
echo "--- 收藏列表 ---"
R=$(api "$BASE/api/favorites" -H "$AUTH")
assert_code "GET /api/favorites" "0" "$(code "$R")"

echo ""
echo "--- 签到 ---"
R=$(api -X POST "$BASE/api/points/sign" -H "$AUTH")
assert_code "POST /api/points/sign" "0" "$(code "$R")"

echo ""
echo "--- 领优惠券 ---"
R=$(api -X POST "$BASE/api/coupons/1/receive" -H "$AUTH")
assert_code "POST /api/coupons/:id/receive" "0" "$(code "$R")"

echo ""
echo "--- 我的优惠券 ---"
R=$(api "$BASE/api/my/coupons" -H "$AUTH")
assert_code "GET /api/my/coupons" "0" "$(code "$R")"

echo ""
echo "--- 用户信息 ---"
R=$(api "$BASE/api/user/profile" -H "$AUTH")
assert_code "GET /api/user/profile" "0" "$(code "$R")"

echo ""
echo "--- 消息列表 ---"
R=$(api "$BASE/api/messages" -H "$AUTH")
assert_code "GET /api/messages" "0" "$(code "$R")"

echo ""
echo "--- ShopXO 兼容层 ---"
R=$(api "$BASE/api.php?s=index/index")
assert_code "ShopXO index/index" "0" "$(code "$R")"
R=$(api "$BASE/api.php?s=goods/category")
assert_code "ShopXO goods/category" "0" "$(code "$R")"
R=$(api "$BASE/api.php?s=goods/detail&goods_id=1")
assert_code "ShopXO goods/detail" "0" "$(code "$R")"
R=$(api "$BASE/api.php?s=base/common")
assert_code "ShopXO base/common" "0" "$(code "$R")"

echo ""
echo "=========================================="
echo "  通过: $PASS  失败: $FAIL"
echo "=========================================="

[ "$FAIL" -eq 0 ] && exit 0 || exit 1
