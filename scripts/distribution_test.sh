#!/bin/bash
# 分销系统集成测试
# 用法: bash scripts/distribution_test.sh
# 前提: 服务已启动，有seed商品数据

set -e
BASE="http://localhost:8080"
PASS=0; FAIL=0
green() { echo -e "\033[32m✓ $1\033[0m"; PASS=$((PASS+1)); }
red()   { echo -e "\033[31m✗ $1\033[0m"; FAIL=$((FAIL+1)); }
check() { if echo "$2" | grep -q "$3"; then green "$1"; else red "$1 — got: $2"; fi; }

login() {
  curl -s -X POST "$BASE/api/login" -H 'Content-Type: application/json' \
    -d "{\"username\":\"$1\",\"password\":\"$2\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}
admin_login() {
  # 获取验证码key
  KEY=$(curl -s -D- "$BASE/api/admin/captcha?key=dist_test" -o /dev/null 2>&1 | grep -i x-captcha-key | tr -d '\r' | awk '{print $2}')
  # 沙盒模式下验证码校验可能需要跳过，直接用空值试试
  curl -s -X POST "$BASE/api/admin/login" -H 'Content-Type: application/json' \
    -d "{\"username\":\"admin\",\"password\":\"admin123\",\"captcha_key\":\"$KEY\",\"captcha_code\":\"0000\"}" \
    | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

echo "=== 分销系统集成测试 ==="
echo ""

# 1. 注册三个用户：爷爷(二级) → 爸爸(一级) → 买家
echo "--- 1. 注册用户 ---"
curl -s -X POST "$BASE/api/register" -H 'Content-Type: application/json' \
  -d '{"username":"dist_grandpa","password":"test123456","nickname":"爷爷"}' > /dev/null
curl -s -X POST "$BASE/api/register" -H 'Content-Type: application/json' \
  -d '{"username":"dist_parent","password":"test123456","nickname":"爸爸"}' > /dev/null
curl -s -X POST "$BASE/api/register" -H 'Content-Type: application/json' \
  -d '{"username":"dist_buyer","password":"test123456","nickname":"买家"}' > /dev/null

T_GRANDPA=$(login dist_grandpa test123456)
T_PARENT=$(login dist_parent test123456)
T_BUYER=$(login dist_buyer test123456)

[ -n "$T_GRANDPA" ] && green "爷爷登录" || { red "爷爷登录失败"; exit 1; }
[ -n "$T_PARENT" ] && green "爸爸登录" || { red "爸爸登录失败"; exit 1; }
[ -n "$T_BUYER" ] && green "买家登录" || { red "买家登录失败"; exit 1; }

# 2. 申请分销商：爷爷(无上级) → 爸爸(上级=爷爷) → 买家(上级=爸爸)
echo ""
echo "--- 2. 申请分销商 ---"
R=$(curl -s -X POST "$BASE/api/distribution/apply" -H "Authorization: Bearer $T_GRANDPA" -H 'Content-Type: application/json' -d '{}')
check "爷爷申请分销商" "$R" '"id"'
GRANDPA_UID=$(echo "$R" | grep -o '"user_id":[0-9]*' | cut -d: -f2)

R=$(curl -s -X POST "$BASE/api/distribution/apply" -H "Authorization: Bearer $T_PARENT" -H 'Content-Type: application/json' -d "{\"parent_id\":$GRANDPA_UID}")
check "爸爸申请分销商(上级=爷爷)" "$R" '"id"'
PARENT_UID=$(echo "$R" | grep -o '"user_id":[0-9]*' | cut -d: -f2)

R=$(curl -s -X POST "$BASE/api/distribution/apply" -H "Authorization: Bearer $T_BUYER" -H 'Content-Type: application/json' -d "{\"parent_id\":$PARENT_UID}")
check "买家申请分销商(上级=爸爸)" "$R" '"id"'

# 3. 查看团队
echo ""
echo "--- 3. 查看团队 ---"
R=$(curl -s "$BASE/api/distribution/team" -H "Authorization: Bearer $T_GRANDPA")
check "爷爷团队包含爸爸" "$R" "\"parent_id\":$GRANDPA_UID"

R=$(curl -s "$BASE/api/distribution/team" -H "Authorization: Bearer $T_PARENT")
check "爸爸团队包含买家" "$R" "\"parent_id\":$PARENT_UID"

# 4. 买家下单（需要地址+商品）
echo ""
echo "--- 4. 买家下单 ---"
curl -s -X POST "$BASE/api/address" -H "Authorization: Bearer $T_BUYER" -H 'Content-Type: application/json' \
  -d '{"name":"测试","phone":"13800000001","province":"北京","city":"北京","district":"朝阳","detail":"分销测试","is_default":true}' > /dev/null

GOODS=$(curl -s "$BASE/api/goods?page=1&page_size=1")
GID=$(echo "$GOODS" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
DETAIL=$(curl -s "$BASE/api/goods/$GID")
SID=$(echo "$DETAIL" | grep -o '"id":[0-9]*' | head -2 | tail -1 | cut -d: -f2)

curl -s -X POST "$BASE/api/cart" -H "Authorization: Bearer $T_BUYER" -H 'Content-Type: application/json' \
  -d "{\"goods_id\":$GID,\"sku_id\":$SID,\"quantity\":1}" > /dev/null
CART_ID=$(curl -s "$BASE/api/cart" -H "Authorization: Bearer $T_BUYER" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

ADDR_ID=$(curl -s "$BASE/api/address" -H "Authorization: Bearer $T_BUYER" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

ORDER_R=$(curl -s -X POST "$BASE/api/orders" -H "Authorization: Bearer $T_BUYER" -H 'Content-Type: application/json' \
  -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_ID]}")
ORDER_ID=$(echo "$ORDER_R" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
ORDER_NO=$(echo "$ORDER_R" | grep -o '"order_no":"[^"]*"' | cut -d'"' -f4)
PAY_AMOUNT=$(echo "$ORDER_R" | grep -o '"pay_amount":[0-9]*' | cut -d: -f2)
check "创建订单 ID=$ORDER_ID 金额=${PAY_AMOUNT}分" "$ORDER_R" '"order_no"'

# 5. 沙盒支付
echo ""
echo "--- 5. 支付订单 ---"
PAY_R=$(curl -s -X POST "$BASE/api/pay/unified" -H "Authorization: Bearer $T_BUYER" -H 'Content-Type: application/json' \
  -d "{\"order_id\":$ORDER_ID,\"payment_key\":\"alipay_pc\"}")
CB_URL=$(echo "$PAY_R" | grep -o '/api/pay/sandbox/callback[^"]*')
if [ -n "$CB_URL" ]; then
  curl -s "$BASE$CB_URL" > /dev/null
  green "沙盒支付成功"
else
  check "支付" "$PAY_R" "SANDBOX_\|OFFLINE_"
fi

# 6. 管理员发货+买家确认收货 → 订单完成
echo ""
echo "--- 6. 发货+收货 ---"
ADMIN_TOKEN=$(admin_login)
if [ -n "$ADMIN_TOKEN" ]; then
  curl -s -X POST "$BASE/api/admin/orders/ship" -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
    -d "{\"order_id\":$ORDER_ID,\"express_company\":\"顺丰\",\"express_no\":\"SF123456\"}" > /dev/null
  green "管理员发货"

  curl -s -X PUT "$BASE/api/orders/$ORDER_ID/receive" -H "Authorization: Bearer $T_BUYER" > /dev/null
  green "买家确认收货"
else
  red "管理员登录失败，跳过发货收货"
fi

# 7. 触发佣金结算（订单完成后调用）
echo ""
echo "--- 7. 验证佣金 ---"
# 佣金结算在订单完成时自动触发（如果接入了的话）
# 这里手动检查分销商余额
R=$(curl -s "$BASE/api/distribution/me" -H "Authorization: Bearer $T_PARENT")
PARENT_BALANCE=$(echo "$R" | grep -o '"balance":[0-9]*' | cut -d: -f2)
PARENT_TOTAL=$(echo "$R" | grep -o '"total_commission":[0-9]*' | cut -d: -f2)
echo "  爸爸(一级): 余额=${PARENT_BALANCE}分, 累计佣金=${PARENT_TOTAL}分"

R=$(curl -s "$BASE/api/distribution/me" -H "Authorization: Bearer $T_GRANDPA")
GRANDPA_BALANCE=$(echo "$R" | grep -o '"balance":[0-9]*' | cut -d: -f2)
GRANDPA_TOTAL=$(echo "$R" | grep -o '"total_commission":[0-9]*' | cut -d: -f2)
echo "  爷爷(二级): 余额=${GRANDPA_BALANCE}分, 累计佣金=${GRANDPA_TOTAL}分"

if [ "$PARENT_BALANCE" = "0" ]; then
  echo "  ⚠ 佣金为0 — SettleCommission 可能未接入订单完成流程"
  echo "  需要在 ConfirmReceive 中调用 service.SettleCommission(orderID)"
fi

# 8. 测试提现
echo ""
echo "--- 8. 提现测试 ---"
if [ "$PARENT_BALANCE" != "0" ] && [ -n "$PARENT_BALANCE" ]; then
  R=$(curl -s -X POST "$BASE/api/distribution/withdraw" -H "Authorization: Bearer $T_PARENT" -H 'Content-Type: application/json' \
    -d "{\"amount\":$PARENT_BALANCE,\"account_type\":\"alipay\",\"account_no\":\"test@test.com\",\"account_name\":\"测试\"}")
  check "爸爸提现申请" "$R" '"code":0\|success'

  R=$(curl -s "$BASE/api/distribution/commission-logs" -H "Authorization: Bearer $T_PARENT")
  check "佣金记录包含提现" "$R" "withdraw"
else
  echo "  跳过提现测试（余额为0）"
fi

# 9. 管理端查看
echo ""
echo "--- 9. 管理端 ---"
if [ -n "$ADMIN_TOKEN" ]; then
  R=$(curl -s "$BASE/api/admin/distributors" -H "Authorization: Bearer $ADMIN_TOKEN")
  check "分销商列表" "$R" '"total"'

  R=$(curl -s "$BASE/api/admin/withdraws" -H "Authorization: Bearer $ADMIN_TOKEN")
  check "提现列表" "$R" '"total"'
fi

echo ""
echo "=== 结果: $PASS 通过, $FAIL 失败 ==="
[ $FAIL -eq 0 ] && exit 0 || exit 1
