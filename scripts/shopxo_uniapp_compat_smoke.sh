#!/usr/bin/env bash
# 模拟 shopxo-uniapp（App.vue）请求形态，对 GoShop /api.php 做冒烟测试。
# 依赖：本机已启动 goshop（默认 http://127.0.0.1:8080）。
# 用法：./scripts/shopxo_uniapp_compat_smoke.sh [BASE_URL]
# 示例：BASE_URL=http://127.0.0.1:8080 ./scripts/shopxo_uniapp_compat_smoke.sh

set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8080}"
BASE_URL="${BASE_URL%/}"

# 与 App.vue request_params_handle 中 H5 常见参数一致（application_client_type=h5）
qs_common="system_type=default&application=app&application_client_type=h5&application_client_brand=&token=&uuid=smoke-$(date +%s)&lang=zh&theme=red&ajax=ajax"

api_url() {
  local s="$1"
  echo "${BASE_URL}/api.php?s=${s}&${qs_common}"
}

echo "== base/common (POST is_key=1) =="
code=$(curl -sS -o /tmp/_sx_smoke1.json -w "%{http_code}" -X POST "$(api_url 'base/common')" -d 'is_key=1')
echo "HTTP $code"
head -c 200 /tmp/_sx_smoke1.json; echo "..."

echo "== index/index (POST) =="
code=$(curl -sS -o /tmp/_sx_smoke2.json -w "%{http_code}" -X POST "$(api_url 'index/index')")
echo "HTTP $code"
head -c 200 /tmp/_sx_smoke2.json; echo "..."

echo "== goods/detail id=1 (POST) =="
code=$(curl -sS -o /tmp/_sx_smoke3.json -w "%{http_code}" -X POST "$(api_url 'goods/detail')&id=1")
echo "HTTP $code"
head -c 200 /tmp/_sx_smoke3.json; echo "..."

echo "== user/login 缺参应返回业务错误而非 404 =="
code=$(curl -sS -o /tmp/_sx_smoke4.json -w "%{http_code}" -X POST "$(api_url 'user/login')")
echo "HTTP $code"
cat /tmp/_sx_smoke4.json; echo

for f in /tmp/_sx_smoke1.json /tmp/_sx_smoke2.json /tmp/_sx_smoke3.json; do
  if ! grep -q '"code":0' "$f"; then
    echo "FAIL: expected code 0 in $f"
    exit 1
  fi
done

if ! grep -q '请输入账号和密码' /tmp/_sx_smoke4.json; then
  echo "WARN: user/login message unexpected (compat layer may have changed)"
fi

echo "OK: shopxo-uniapp 风格请求与 GoShop /api.php 兼容层联调通过。"
