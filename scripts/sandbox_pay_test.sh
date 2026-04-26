#!/bin/bash
# 支付沙盒测试脚本
# 前提：config.yaml 中 payment.sandbox: true，服务已启动
#
# 用法: bash scripts/sandbox_pay_test.sh

set -e
BASE="http://localhost:8080"
PASS=0
FAIL=0

green() { echo -e "\033[32m✓ $1\033[0m"; PASS=$((PASS+1)); }
red()   { echo -e "\033[31m✗ $1\033[0m"; FAIL=$((FAIL+1)); }
check() {
  if echo "$2" | grep -q "$3"; then green "$1"; else red "$1 (expected '$3', got: $2)"; fi
}

echo "=== 支付沙盒测试 ==="
echo ""

# 1. 注册测试用户
echo "--- 准备测试数据 ---"
REG=$(curl -s -X POST "$BASE/api/register" -H 'Content-Type: application/json' \
  -d '{"username":"sandbox_tester","password":"test123456","nickname":"沙盒测试"}')
echo "注册: $REG"

# 2. 登录
LOGIN=$(curl -s -X POST "$BASE/api/login" -H 'Content-Type: application/json' \
  -d '{"username":"sandbox_tester","password":"test123456"}')
TOKEN=$(echo "$LOGIN" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  red "登录失败: $LOGIN"
  exit 1
fi
green "登录成功"
AUTH="Authorization: Bearer $TOKEN"

# 3. 添加收货地址
ADDR=$(curl -s -X POST "$BASE/api/address" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"name":"测试","phone":"13800138000","province":"北京","city":"北京","district":"朝阳","detail":"测试地址","is_default":true}')
ADDR_ID=$(echo "$ADDR" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
green "创建地址 ID=$ADDR_ID"

# 4. 获取一个商品和SKU
GOODS=$(curl -s "$BASE/api/goods?page=1&page_size=1")
GOODS_ID=$(echo "$GOODS" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
if [ -z "$GOODS_ID" ]; then
  red "无商品数据，请先初始化seed"
  exit 1
fi
DETAIL=$(curl -s "$BASE/api/goods/$GOODS_ID")
SKU_ID=$(echo "$DETAIL" | grep -o '"id":[0-9]*' | head -2 | tail -1 | cut -d: -f2)
green "商品 ID=$GOODS_ID, SKU ID=$SKU_ID"

# 5. 加入购物车
CART=$(curl -s -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"goods_id\":$GOODS_ID,\"sku_id\":$SKU_ID,\"quantity\":1}")
CART_ID=$(echo "$CART" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
green "加入购物车 ID=$CART_ID"

# 测试所有支付方式
PAYMENT_KEYS="wechat_jsapi wechat_h5 wechat_app wechat_native alipay_pc alipay_h5 alipay_app alipay_mini offline"

echo ""
echo "--- 逐一测试支付方式 ---"

for PAY_KEY in $PAYMENT_KEYS; do
  # 创建订单
  ORDER=$(curl -s -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_ID]}")
  ORDER_ID=$(echo "$ORDER" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

  if [ -z "$ORDER_ID" ]; then
    # 购物车可能已清空，重新加
    CART=$(curl -s -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' \
      -d "{\"goods_id\":$GOODS_ID,\"sku_id\":$SKU_ID,\"quantity\":1}")
    CART_ID=$(echo "$CART" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
    ORDER=$(curl -s -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' \
      -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_ID]}")
    ORDER_ID=$(echo "$ORDER" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
  fi

  if [ -z "$ORDER_ID" ]; then
    red "[$PAY_KEY] 创建订单失败"
    continue
  fi

  # 发起支付
  PAY_RESP=$(curl -s -X POST "$BASE/api/pay/unified" -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"order_id\":$ORDER_ID,\"payment_key\":\"$PAY_KEY\"}")

  # offline 和 wallet 直接完成，其他走沙盒回调
  if [ "$PAY_KEY" = "offline" ]; then
    check "[$PAY_KEY] 支付" "$PAY_RESP" "OFFLINE_"
  else
    # 提取沙盒回调URL
    CALLBACK_URL=$(echo "$PAY_RESP" | grep -o '/api/pay/sandbox/callback[^"]*' | head -1)
    if [ -z "$CALLBACK_URL" ]; then
      check "[$PAY_KEY] 支付" "$PAY_RESP" "SANDBOX_"
    else
      # 触发沙盒回调
      CB_RESP=$(curl -s "$BASE$CALLBACK_URL")
      check "[$PAY_KEY] 支付+回调" "$CB_RESP" "沙盒支付成功"
    fi
  fi
done

# 测试钱包支付（需要余额）
echo ""
echo "--- 钱包支付测试 ---"
CART=$(curl -s -X POST "$BASE/api/cart" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"goods_id\":$GOODS_ID,\"sku_id\":$SKU_ID,\"quantity\":1}")
CART_ID=$(echo "$CART" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
ORDER=$(curl -s -X POST "$BASE/api/orders" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"address_id\":$ADDR_ID,\"cart_ids\":[$CART_ID]}")
ORDER_ID=$(echo "$ORDER" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
WALLET_RESP=$(curl -s -X POST "$BASE/api/pay/unified" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"order_id\":$ORDER_ID,\"payment_key\":\"wallet\"}")
# 钱包余额不足是预期的
if echo "$WALLET_RESP" | grep -q "余额不足\|WALLET_"; then
  green "[wallet] 钱包支付（余额不足或成功均正常）"
else
  red "[wallet] 钱包支付: $WALLET_RESP"
fi

echo ""
echo "=== 结果: $PASS 通过, $FAIL 失败 ==="
[ $FAIL -eq 0 ] && exit 0 || exit 1
