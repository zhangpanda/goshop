#!/bin/bash
# GoShop 核心流程集成测试
# 需要后端服务运行在 localhost:8080（或 BASE 指向的地址）
# 用法: ./scripts/integration_test.sh
# 环境变量:
#   BASE — 默认 http://localhost:8080
#   GOSHOP_PAYMENT_SANDBOX=1 — 且 config payment.sandbox=true 时，额外跑「ShopXO 多订单 + 沙盒回调」

set -e
BASE="${BASE:-http://localhost:8080}"
PASS=0
FAIL=0

# 超时与失败时非零退出，避免挂死
CURL=(curl -sS --connect-timeout 8 --max-time 60)

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

api() { "${CURL[@]}" "$@"; }
code() { echo "$1" | python3 -c "import sys,json;print(json.load(sys.stdin)['code'])"; }

echo "========== GoShop 集成测试 =========="
echo "BASE=$BASE"

if ! "${CURL[@]}" -o /dev/null -f "$BASE/api/site-config"; then
  echo "❌ 无法访问 $BASE/api/site-config（请先启动服务，或设置 BASE）"
  exit 1
fi

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
ADDR_ID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")

echo ""
echo "--- 创建订单 ---"
R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_ID]}")
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
echo "--- 支付方式列表（渠道 key 覆盖） ---"
R=$(api "$BASE/api/payments")
assert_code "GET /api/payments" "0" "$(code "$R")"
KEYS=$(echo "$R" | python3 -c "import sys,json,re
d=json.load(sys.stdin).get('data')or[]
keys=set()
for p in d:
  cfg=p.get('config')or''
  m=re.search(r'\"payment_key\"\\s*:\\s*\"([^\"]+)\"',cfg)
  if m: keys.add(m.group(1))
need={'offline','wallet','wechat_jsapi','wechat_h5','wechat_app','wechat_native','alipay_h5','alipay_pc','alipay_app','alipay_mini','alipay_face','paypal'}
missing=need-keys
if missing: print('MISSING:'+','.join(sorted(missing))); sys.exit(1)
print('ok')")
if [ "$KEYS" != "ok" ]; then
  echo "  ❌ 缺支付方式 payment_key: $KEYS（新库应含 EnsureDefaultPayments 全渠道）"
  FAIL=$((FAIL+1))
else
  echo "  ✅ GET /api/payments 含默认全渠道 payment_key"
  PASS=$((PASS+1))
fi

echo ""
echo "--- REST /api/pay/unified 线下支付（待付款订单） ---"
R=$(api "$BASE/api/payments")
PAY_OFF_UNI=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin);pid=0
for p in d.get('data')or[]:
  if'\"payment_key\":\"offline\"'in(p.get('config')or'')or'offline'in(p.get('config')or''):pid=p['id'];break
print(pid)")
if [ "$PAY_OFF_UNI" = "0" ]; then
  echo "  ❌ 无 offline 支付方式"
  FAIL=$((FAIL+1))
else
  R=$(api -X POST "$BASE/api/pay/unified" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"order_id\":$ORDER_ID,\"payment_key\":\"offline\",\"payment_id\":$PAY_OFF_UNI}")
  assert_code "POST /api/pay/unified offline" "0" "$(code "$R")"
  R=$(api "$BASE/api/orders/$ORDER_ID" -H "$AUTH")
  ST_PAID=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['status'])")
  assert_code "线下支付后订单已付(status=1)" "1" "$ST_PAID"
fi

echo ""
echo "--- 另建待付款订单并取消 ---"
R=$(api -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' -d '{"goods_id":1,"sku_id":1,"quantity":1}')
CART_CANCEL=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(d['id']if isinstance(d,dict)else d[0]['id'])")
R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_CANCEL]}")
ORDER_CANCEL=$(echo "$R" | python3 -c "import sys,json;x=json.load(sys.stdin)['data'];print(x[0]['id']if isinstance(x,list)else x['id'])")
R=$(api -X PUT "$BASE/api/orders/$ORDER_CANCEL/cancel" -H "$AUTH")
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
echo "--- ShopXO order/pay 多订单线下支付 ---"
R=$(api "$BASE/api/payments")
PAY_OFFLINE=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin);pid=0
for p in d.get('data')or[]:
  if'offline'in(p.get('config')or''):pid=p['id'];break
print(pid)")
if [ "$PAY_OFFLINE" = "0" ]; then
  echo "  ❌ 无线下支付方式（需 EnsureDefaultPayments）"
  FAIL=$((FAIL+1))
else
  R=$(api -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' -d '{"goods_id":1,"sku_id":1,"quantity":1}')
  CART_M1=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(d['id']if isinstance(d,dict)else d[0]['id'])")
  R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_M1]}")
  O_M1=$(echo "$R" | python3 -c "import sys,json;x=json.load(sys.stdin)['data'];print(x[0]['id']if isinstance(x,list)else x['id'])")
  R=$(api -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' -d '{"goods_id":1,"sku_id":1,"quantity":1}')
  CART_M2=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(d['id']if isinstance(d,dict)else d[0]['id'])")
  R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_M2]}")
  O_M2=$(echo "$R" | python3 -c "import sys,json;x=json.load(sys.stdin)['data'];print(x[0]['id']if isinstance(x,list)else x['id'])")
  R=$(api -X POST "$BASE/api.php?s=order/pay&token=${TOKEN}" -H 'Content-Type: application/json' -d "{\"ids\":\"${O_M1},${O_M2}\",\"payment_id\":${PAY_OFFLINE}}")
  assert_code "ShopXO order/pay 多单 offline" "0" "$(code "$R")"
  R=$(api "$BASE/api/orders/${O_M1}" -H "$AUTH")
  ST1=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['status'])")
  assert_code "多单支付后订单1已付(status=1)" "1" "$ST1"
  R=$(api "$BASE/api/orders/${O_M2}" -H "$AUTH")
  ST2=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['status'])")
  assert_code "多单支付后订单2已付(status=1)" "1" "$ST2"
fi

if [ "${GOSHOP_PAYMENT_SANDBOX:-}" = "1" ]; then
  echo ""
  echo "--- ShopXO order/pay 多订单 + 沙盒回调 ---"
  R=$(api "$BASE/api/payments")
  PAY_WX=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin);pid=0
for p in d.get('data')or[]:
  if'wechat_jsapi'in(p.get('config')or''):pid=p['id'];break
print(pid)")
  if [ "$PAY_WX" = "0" ]; then
    echo "  ❌ 无 wechat_jsapi 支付方式"
    FAIL=$((FAIL+1))
  else
    R=$(api -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' -d '{"goods_id":1,"sku_id":1,"quantity":1}')
    CART_S1=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(d['id']if isinstance(d,dict)else d[0]['id'])")
    R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_S1]}")
    O_S1=$(echo "$R" | python3 -c "import sys,json;x=json.load(sys.stdin)['data'];print(x[0]['id']if isinstance(x,list)else x['id'])")
    R=$(api -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' -d '{"goods_id":1,"sku_id":1,"quantity":1}')
    CART_S2=$(echo "$R" | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(d['id']if isinstance(d,dict)else d[0]['id'])")
    R=$(api -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_S2]}")
    O_S2=$(echo "$R" | python3 -c "import sys,json;x=json.load(sys.stdin)['data'];print(x[0]['id']if isinstance(x,list)else x['id'])")
    R=$(api -X POST "$BASE/api.php?s=order/pay&token=${TOKEN}" -H 'Content-Type: application/json' -d "{\"ids\":\"${O_S1},${O_S2}\",\"payment_id\":${PAY_WX}}")
    assert_code "ShopXO order/pay 多单 wechat(沙盒)" "0" "$(code "$R")"
    CB=$(echo "$R" | python3 -c "import sys,json
d=json.load(sys.stdin).get('data')or{}
x=d.get('data')
print(x if isinstance(x,str)and x.startswith('/')else'')")
    if [ -z "$CB" ]; then
      echo "  ❌ 响应中无沙盒回调路径"
      FAIL=$((FAIL+1))
    else
      R=$(api "$BASE$CB")
      assert_code "沙盒回调 PayLog 多单" "0" "$(code "$R")"
      R=$(api "$BASE/api/orders/${O_S1}" -H "$AUTH")
      SS1=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['status'])")
      assert_code "沙盒后订单1已付" "1" "$SS1"
      R=$(api "$BASE/api/orders/${O_S2}" -H "$AUTH")
      SS2=$(echo "$R" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['status'])")
      assert_code "沙盒后订单2已付" "1" "$SS2"
    fi
  fi
fi

echo ""
echo "=========================================="
echo "  通过: $PASS  失败: $FAIL"
echo "=========================================="

[ "$FAIL" -eq 0 ] && exit 0 || exit 1
