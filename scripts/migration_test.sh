#!/bin/bash
# ShopXO → GoShop 迁移自动化验证（自包含，不依赖 shopxo.sql）
# MySQL 口令：默认 goshop123（与仓库 docker-compose.yml 中 MYSQL_ROOT_PASSWORD 一致）；
# 本地库若不同，请设置环境变量 GOSHOP_MIGRATION_MYSQL_PASSWORD。
set -euo pipefail
cd "$(dirname "$0")/.."

MIG_DB_PASS="${GOSHOP_MIGRATION_MYSQL_PASSWORD:-goshop123}"
MIG_DB_HOST="${GOSHOP_MIGRATION_MYSQL_HOST:-127.0.0.1}"
MIG_DB_PORT="${GOSHOP_MIGRATION_MYSQL_PORT:-3306}"

SRC=shopxo_mig_test_src
DST=shopxo_mig_test_dst

MY_CNF=$(mktemp)
chmod 600 "$MY_CNF"
cat >"$MY_CNF" <<EOF
[client]
user=root
password=${MIG_DB_PASS}
host=${MIG_DB_HOST}
port=${MIG_DB_PORT}
default-character-set=utf8mb4
EOF

mysql_cli() {
  mysql --defaults-extra-file="$MY_CNF" "$@"
}

cleanup() {
  rm -f "$MY_CNF" 2>/dev/null || true
  if [ -f config.yaml.bak ]; then
    cp -f config.yaml.bak config.yaml
    rm -f config.yaml.bak
  fi
}
trap cleanup EXIT

apply_test_config() {
  local dbname="$1"
  awk -v p="$MIG_DB_PASS" -v db="$dbname" -v h="$MIG_DB_HOST" -v pt="$MIG_DB_PORT" '
    /^db:/ { indb=1; inr=0; print; next }
    /^redis:/ { indb=0; inr=1; print; next }
    /^[^[:space:]#]/ { indb=0; inr=0 }
    indb && /^  password:/ { print "  password: \"" p "\""; next }
    indb && /^  dbname:/ { print "  dbname: " db; next }
    indb && /^  host:/ { print "  host: " h; next }
    indb && /^  port:/ { print "  port: " pt; next }
    inr && /^  host:/ { print "  host: \"\""; next }
    { print }
  ' config.yaml.bak > config.yaml
}

cp config.yaml config.yaml.bak

if [ ! -x ./bin/goshop ]; then
  echo "构建 ./bin/goshop …"
  go build -o bin/goshop ./cmd/server/main.go
fi

echo "=== 1. 创建测试库 ==="
mysql_cli -e "DROP DATABASE IF EXISTS $SRC; DROP DATABASE IF EXISTS $DST;"
mysql_cli -e "CREATE DATABASE $SRC CHARACTER SET utf8mb4; CREATE DATABASE $DST CHARACTER SET utf8mb4;"

echo "=== 2. 创建 ShopXO 源表 + 模拟数据 ==="
mysql_cli "$SRC" <<'SQL'
SET NAMES utf8mb4;
CREATE TABLE sxo_user (id int unsigned AUTO_INCREMENT PRIMARY KEY, username varchar(60) DEFAULT '', nickname varchar(60) DEFAULT '', mobile varchar(30) DEFAULT '', avatar varchar(255) DEFAULT '', integral int DEFAULT 0, locking_integral int DEFAULT 0, status tinyint DEFAULT 0, is_delete_time int DEFAULT 0, add_time int DEFAULT 0, upd_time int DEFAULT 0);
CREATE TABLE sxo_goods_category (id int unsigned AUTO_INCREMENT PRIMARY KEY, pid int DEFAULT 0, name varchar(60) DEFAULT '', icon varchar(255) DEFAULT '', sort int DEFAULT 0, is_enable tinyint DEFAULT 1, add_time int DEFAULT 0, upd_time int DEFAULT 0);
CREATE TABLE sxo_goods (id int unsigned AUTO_INCREMENT PRIMARY KEY, brand_id int DEFAULT 0, title varchar(160) DEFAULT '', simple_desc varchar(230) DEFAULT '', images varchar(255) DEFAULT '', content_web text, is_shelves tinyint DEFAULT 0, sort_level int DEFAULT 0, min_price decimal(10,2) DEFAULT 0, sales_count int DEFAULT 0, access_count int DEFAULT 0, inventory int DEFAULT 0, coding varchar(180) DEFAULT '', give_integral int DEFAULT 0, is_delete_time int DEFAULT 0, add_time int DEFAULT 0, upd_time int DEFAULT 0);
CREATE TABLE sxo_goods_category_join (goods_id int DEFAULT 0, category_id int DEFAULT 0);
CREATE TABLE sxo_goods_photo (id int unsigned AUTO_INCREMENT PRIMARY KEY, goods_id int DEFAULT 0, images varchar(255) DEFAULT '', sort int DEFAULT 0, is_show tinyint DEFAULT 1, add_time int DEFAULT 0);
CREATE TABLE sxo_goods_spec_base (id int unsigned AUTO_INCREMENT PRIMARY KEY, goods_id int DEFAULT 0, price decimal(10,2) DEFAULT 0, inventory int DEFAULT 0, coding varchar(180) DEFAULT '', add_time int DEFAULT 0);
CREATE TABLE sxo_goods_spec_value (id int unsigned AUTO_INCREMENT PRIMARY KEY, goods_id int DEFAULT 0, goods_spec_base_id int DEFAULT 0, value varchar(120) DEFAULT '', add_time int DEFAULT 0);
CREATE TABLE sxo_region (id int unsigned PRIMARY KEY, pid int DEFAULT 0, name varchar(60) DEFAULT '', level tinyint DEFAULT 0, sort int DEFAULT 0, is_enable tinyint DEFAULT 1, add_time int DEFAULT 0);
CREATE TABLE sxo_user_address (id int unsigned AUTO_INCREMENT PRIMARY KEY, user_id int DEFAULT 0, name varchar(60) DEFAULT '', tel varchar(30) DEFAULT '', province int DEFAULT 0, city int DEFAULT 0, county int DEFAULT 0, address varchar(255) DEFAULT '', is_default tinyint DEFAULT 0, add_time int DEFAULT 0, upd_time int DEFAULT 0);
CREATE TABLE sxo_order (id int unsigned AUTO_INCREMENT PRIMARY KEY, order_no varchar(60) DEFAULT '', user_id int DEFAULT 0, total_price decimal(10,2) DEFAULT 0, pay_price decimal(10,2) DEFAULT 0, status tinyint DEFAULT 0, pay_status tinyint DEFAULT 0, user_note varchar(255) DEFAULT '', payment_id int DEFAULT 0, order_model tinyint DEFAULT 0, is_delete_time int DEFAULT 0, add_time int DEFAULT 0, upd_time int DEFAULT 0, pay_time int DEFAULT 0, delivery_time int DEFAULT 0, collect_time int DEFAULT 0);
CREATE TABLE sxo_order_address (id int unsigned AUTO_INCREMENT PRIMARY KEY, order_id int DEFAULT 0, name varchar(60) DEFAULT '', tel varchar(30) DEFAULT '', province_name varchar(60) DEFAULT '', city_name varchar(60) DEFAULT '', county_name varchar(60) DEFAULT '', address varchar(255) DEFAULT '', add_time int DEFAULT 0);
CREATE TABLE sxo_order_detail (id int unsigned AUTO_INCREMENT PRIMARY KEY, order_id int DEFAULT 0, goods_id int DEFAULT 0, title varchar(160) DEFAULT '', images varchar(255) DEFAULT '', spec varchar(200) DEFAULT '', spec_coding varchar(180) DEFAULT '', spec_desc varchar(200) DEFAULT '', price decimal(10,2) DEFAULT 0, buy_number int DEFAULT 0, add_time int DEFAULT 0);
CREATE TABLE sxo_brand (id int unsigned AUTO_INCREMENT PRIMARY KEY, name varchar(60) DEFAULT '', logo varchar(255) DEFAULT '', sort int DEFAULT 0, is_enable tinyint DEFAULT 1, add_time int DEFAULT 0, upd_time int DEFAULT 0);
CREATE TABLE sxo_article (id int unsigned AUTO_INCREMENT PRIMARY KEY, title varchar(160) DEFAULT '', content text, article_category_id int DEFAULT 0, cover varchar(255) DEFAULT '', access_count int DEFAULT 0, is_enable tinyint DEFAULT 1, add_time int DEFAULT 0, upd_time int DEFAULT 0);
CREATE TABLE sxo_admin (id int unsigned AUTO_INCREMENT PRIMARY KEY, username varchar(60) DEFAULT '', mobile varchar(30) DEFAULT '', role_id int DEFAULT 1, status tinyint DEFAULT 0, add_time int DEFAULT 0, upd_time int DEFAULT 0);

INSERT INTO sxo_user VALUES (1,'testuser','测试用户','13800001111','/avatar.jpg',100,0,0,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP()),(2,'vipuser','VIP客户','13900002222','',500,10,0,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
INSERT INTO sxo_goods_category VALUES (1,0,'手机数码','',100,1,UNIX_TIMESTAMP(),UNIX_TIMESTAMP()),(2,0,'服装鞋帽','',90,1,UNIX_TIMESTAMP(),UNIX_TIMESTAMP()),(3,1,'智能手机','',80,1,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
INSERT INTO sxo_goods VALUES (1,0,'iPhone 15 Pro','苹果旗舰','/img/iphone.jpg','<p>详情</p>',1,10,7999.00,50,1000,100,'IP15P',10,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP()),(2,0,'运动T恤','透气速干','/img/tshirt.jpg','<p>T恤</p>',1,5,99.00,200,500,999,'TS001',1,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
INSERT INTO sxo_goods_category_join VALUES (1,1),(1,3),(2,2);
INSERT INTO sxo_goods_photo VALUES (1,1,'/img/iphone-1.jpg',0,1,UNIX_TIMESTAMP()),(2,1,'/img/iphone-2.jpg',1,1,UNIX_TIMESTAMP());
INSERT INTO sxo_goods_spec_base VALUES (1,1,7999.00,50,'IP15P-256',UNIX_TIMESTAMP()),(2,1,8999.00,30,'IP15P-512',UNIX_TIMESTAMP());
INSERT INTO sxo_goods_spec_value VALUES (1,1,1,'256GB',UNIX_TIMESTAMP()),(2,1,2,'512GB',UNIX_TIMESTAMP());
INSERT INTO sxo_region VALUES (110000,0,'北京市',1,0,1,UNIX_TIMESTAMP()),(110100,110000,'北京市',2,0,1,UNIX_TIMESTAMP()),(110101,110100,'东城区',3,0,1,UNIX_TIMESTAMP());
INSERT INTO sxo_user_address VALUES (1,1,'张三','13800001111',110000,110100,110101,'长安街1号',1,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
INSERT INTO sxo_order VALUES (1,'GO202604270001',1,7999.00,7999.00,4,1,'请尽快发货',1,0,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),UNIX_TIMESTAMP()),(2,'GO202604270002',2,198.00,188.00,2,1,'',1,0,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),0,0);
INSERT INTO sxo_order_address VALUES (1,1,'张三','13800001111','北京市','北京市','东城区','长安街1号',UNIX_TIMESTAMP()),(2,2,'VIP客户','13900002222','上海市','上海市','浦东新区','陆家嘴100号',UNIX_TIMESTAMP());
INSERT INTO sxo_order_detail VALUES (1,1,1,'iPhone 15 Pro','/img/iphone.jpg','256GB','IP15P-256','256GB',7999.00,1,UNIX_TIMESTAMP()),(2,2,2,'运动T恤','/img/tshirt.jpg','','','默认',99.00,2,UNIX_TIMESTAMP());
INSERT INTO sxo_brand VALUES (1,'Apple','/brand/apple.png',100,1,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
INSERT INTO sxo_article VALUES (1,'新品发布','<p>iPhone 15 Pro 正式发售</p>',1,'/article/cover.jpg',999,1,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
INSERT INTO sxo_admin VALUES (1,'admin','13800000000',1,0,UNIX_TIMESTAMP(),UNIX_TIMESTAMP());
SQL

echo "=== 3. 创建 GoShop 目标表 ==="
kill $(lsof -ti :8080) 2>/dev/null || true
sleep 1
apply_test_config "$DST"
GOSHOP_E2E=1 ./bin/goshop &
PID=$!; sleep 4; kill $PID 2>/dev/null; sleep 1

echo "=== 4. 执行迁移 SQL ==="
mysql_cli "$DST" -e "SET FOREIGN_KEY_CHECKS=0; TRUNCATE users; TRUNCATE categories; TRUNCATE goods; TRUNCATE goods_skus; TRUNCATE orders; TRUNCATE order_items; TRUNCATE addresses; TRUNCATE brands; TRUNCATE articles; TRUNCATE admins; SET FOREIGN_KEY_CHECKS=1;"
go run ./cmd/shopxo-import -host "$MIG_DB_HOST" -port "$MIG_DB_PORT" -user root -password "$MIG_DB_PASS" -from "$SRC" -to "$DST"

# ShopXO 仅迁移 admins，不包含 roles；先 TRUNCATE admins 再导入会留下「无角色」数据。补全超级管理员角色以保证 RBAC。
mysql_cli "$DST" <<'EOSQL'
INSERT INTO roles (id, name, `desc`, powers, status, created_at, updated_at)
VALUES (1, '超级管理员', '拥有所有权限', '["*"]', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  `desc` = VALUES(`desc`),
  powers = VALUES(powers),
  status = VALUES(status);
UPDATE admins SET role_id = 1 WHERE id = 1;
EOSQL

echo "=== 5. 验证数据 ==="
FAIL=0
check() {
  local desc="$1" sql="$2" expect="$3"
  result=$(mysql_cli -N -e "$sql" "$DST" 2>/dev/null | tr -d '[:space:]')
  if [ "$result" = "$expect" ]; then echo "  ✅ $desc"; else echo "  ❌ $desc: 期望=$expect 实际=$result"; FAIL=1; fi
}

check "用户数=2" "SELECT COUNT(*) FROM users" "2"
check "用户1状态=正常" "SELECT status FROM users WHERE id=1" "1"
check "分类数=3" "SELECT COUNT(*) FROM categories" "3"
check "子分类父ID" "SELECT parent_id FROM categories WHERE id=3" "1"
check "商品数=2" "SELECT COUNT(*) FROM goods" "2"
check "商品1分类=1" "SELECT category_id FROM goods WHERE id=1" "1"
check "SKU数=3" "SELECT COUNT(*) FROM goods_skus" "3"
check "SKU1价格=799900分" "SELECT price FROM goods_skus WHERE id=1" "799900"
check "无规格SKU价格=9900分" "SELECT price FROM goods_skus WHERE goods_id=2" "9900"
check "订单数=2" "SELECT COUNT(*) FROM orders" "2"
check "订单1金额=799900分" "SELECT pay_amount FROM orders WHERE id=1" "799900"
check "订单1状态=完成(3)" "SELECT status FROM orders WHERE id=1" "3"
check "订单2状态=待发货(1)" "SELECT status FROM orders WHERE id=2" "1"
check "订单地址含张三" "SELECT JSON_UNQUOTE(JSON_EXTRACT(address,'$.name')) FROM orders WHERE id=1" "张三"
check "明细数=2" "SELECT COUNT(*) FROM order_items" "2"
check "明细1价格=799900分" "SELECT price FROM order_items WHERE id=1" "799900"
check "地址数=1" "SELECT COUNT(*) FROM addresses" "1"
check "地址省=北京市" "SELECT province FROM addresses WHERE id=1" "北京市"
check "品牌=Apple" "SELECT name FROM brands WHERE id=1" "Apple"
check "文章数=1" "SELECT COUNT(*) FROM articles" "1"
check "管理员状态=正常" "SELECT status FROM admins WHERE id=1" "1"

echo ""
echo "=== 6. API 验证 ==="
mysql_cli "$DST" -e "UPDATE admins SET password='\$2a\$10\$Q7LicCzsoEGtbWhumeGrQOzrK5F2/dO2bbjN5De2jEewY8tCxCPgy' WHERE id=1;"
apply_test_config "$DST"
GOSHOP_E2E=1 ./bin/goshop &
PID=$!
READY=0
for _ in $(seq 1 60); do
  if (echo >/dev/tcp/127.0.0.1/8080) &>/dev/null; then
    READY=1
    break
  fi
  sleep 1
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "  ❌ 后端进程已退出"
    FAIL=1
    break
  fi
done
if [ "$READY" != 1 ] && [ "$FAIL" -eq 0 ]; then
  echo "  ❌ 等待 :8080 监听超时"
  FAIL=1
fi
sleep 2

LOGIN=$(curl -s -X POST http://127.0.0.1:8080/api/admin/login -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123","captcha_key":"t","captcha_code":"0"}')
if echo "$LOGIN" | grep -q '"code":0'; then
  TOKEN=$(echo "$LOGIN" | sed 's/.*"token":"\([^"]*\)".*/\1/')
  echo "  ✅ 管理员登录"
  curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/goods | grep -q 'iPhone' && echo "  ✅ 商品API" || { echo "  ❌ 商品API"; FAIL=1; }
  curl -s -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:8080/api/admin/orders?page=1&page_size=20" | grep -q 'order_no' && echo "  ✅ 订单API" || { echo "  ❌ 订单API"; FAIL=1; }
  curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/admin/users | grep -q 'testuser' && echo "  ✅ 用户API" || { echo "  ❌ 用户API"; FAIL=1; }
  if ! CJSON=$(curl -fsS -H 'Accept-Encoding: identity' http://127.0.0.1:8080/api/categories); then
    echo "  ❌ 分类API (curl)"; FAIL=1
  elif ! echo "$CJSON" | python3 -c 'import sys,json; j=json.load(sys.stdin); s=json.dumps(j,ensure_ascii=False); sys.exit(0 if "手机数码" in s else 1)'; then
    echo "  ❌ 分类API"; FAIL=1
  else
    echo "  ✅ 分类API"
  fi
else
  echo "  ❌ 登录失败: $LOGIN"; FAIL=1
fi

kill $PID 2>/dev/null

echo ""
echo "=== 7. 清理测试库 ==="
mysql_cli -e "DROP DATABASE IF EXISTS $SRC; DROP DATABASE IF EXISTS $DST;"
kill $(lsof -ti :8080) 2>/dev/null || true
sleep 1
# 恢复 config 由 trap cleanup 完成；仅当开发者曾用 GOSHOP_E2E 常驻时可选手动重启：
# GOSHOP_E2E=1 nohup ./bin/goshop > /tmp/goshop-e2e.log 2>&1 &

echo ""
[ "$FAIL" -eq 0 ] && echo "🎉 迁移测试全部通过！ShopXO 老用户可正常迁移" || echo "💥 有失败项"
exit "$FAIL"
