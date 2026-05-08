# ShopXO 迁移到 GoShop 指南

## 适用版本与结构依据

- **ShopXO v6.x**（下文 SQL 按 **v6.8** 官方库表核对）
- 表前缀默认 `sxo_`（安装时可自定义，全文以 `sxo_` 为例）
- **表结构依据**：本地可参考 `/tmp/shopxo/config/shopxo.sql`（或官方包内 `config/shopxo.sql`）；执行前仍建议对你方真实库 `SHOW CREATE TABLE` 抽查

> GoShop 为独立实现；迁移仅解决「数据落库」问题，业务规则以后端代码为准。

## 重要说明（执行前必读）

1. **推荐执行顺序**见文末「建议迁移顺序」；违反顺序可能导致外键或逻辑不一致（若你库上启用了外键）。
2. **方案 B（md5 兼容登录）**：当前仓库 `User` 无 `salt`/`old_pwd`，登录未实现 md5 回退，需自研（见「第二步」）。
3. **管理员**：`admins.password` 为 bcrypt；ShopXO 为 `login_pwd`+`login_salt`（md5），须重置密码。**状态码与 ShopXO 相反**：ShopXO `0=正常,1=无效` → GoShop `1=正常,0=禁用`。
4. **订单地址**：ShopXO v6.8 **没有** `sxo_order.address_data` 字段，快照在 **`sxo_order_address`**，须 JOIN 或子查询拼 JSON。
5. **商品分类**：`sxo_goods` **无** `category_id`，分类在 **`sxo_goods_category_join`**（多对多）；GoShop `goods.category_id` 为单值，文中取 **每个商品最小 `category_id`**，若需「主分类」请按业务改子查询。
6. **相册**：`sxo_goods.images` 为**封面单图**；多图在 **`sxo_goods_photo`**。文中用 **`GROUP_CONCAT(JSON_QUOTE(...) ORDER BY …)` + `CAST(… AS JSON)`** 生成有序 JSON 数组（兼容 **MySQL 8/9**；`JSON_ARRAYAGG(expr ORDER BY …)` 在部分环境会报语法错误）。
7. **已退款订单**：ShopXO 用 `pay_status` 等区分部分退款，GoShop 有状态「已退款」。若需精细映射，请在 `sxo_order` 上增加对 `pay_status` 的 `CASE`（文中给出口子）。
8. **中文与 utf8mb4**：源、目标库表与库本身应使用 **utf8mb4**。手工执行下文 SQL、或用 `mysql` 客户端造数时，会话必须是 **utf8mb4**，否则中文会落库成乱码（API 里可见类似 `æ‰‹æœº` 的「双重误读」）。详见下节。

## MySQL 客户端字符集（utf8mb4）

- **`go run ./cmd/shopxo-import`**：连接串已带 **`charset=utf8mb4`**（`internal/shopxomigrate/data.sql` 前几行亦有 `SET NAMES utf8mb4`），一般无需额外设置。
- **命令行 `mysql`**：建议使用配置文件，避免仅依赖终端编码：

```ini
# 例如 ~/.my.cnf 片段（仅示例，注意文件权限）
[client]
default-character-set=utf8mb4
```

或在**每次**执行迁移 SQL 前显式执行：

```sql
SET NAMES utf8mb4;
```

- **自助回归**：`scripts/migration_test.sh` 使用 **`--defaults-extra-file`**，并设置 **`default-character-set=utf8mb4`**、在造源表数据前 **`SET NAMES utf8mb4`**，与上述要求一致。

## 核心差异对照

| 差异项 | ShopXO v6.8 | GoShop |
|--------|-------------|--------|
| 表前缀 | `sxo_` | 无前缀 |
| 金额 | 元（decimal） | 分（int64） |
| 时间 | Unix 时间戳 | `datetime` |
| 用户密码 | `md5(salt + pwd)` | bcrypt |
| 用户状态 | 0正常,1禁言,2禁登 | 0禁用,1正常 |
| 订单 status | 0待确认,1待支付,2待发货,3待收货,4完成,5取消,6关闭 | 0待付款,1待发货,2待收货,3完成,4取消,5退款… |
| 管理员状态 | 0正常,1无效 | 0禁用,1正常 |
| 管理员密码 | `login_pwd` + `login_salt` | `password`(bcrypt) |

## 前置条件

- GoShop 已首次启动并完成自动建表
- 可访问 ShopXO 的 MySQL（**先在测试库验证**）
- 下文默认：`shopxo` = 源库，`goshop` = 目标库；按需替换库名与前缀
- **本机演练**：优先用 **`go run ./cmd/shopxo-import`**（见下节）；亦可空库导入 `config/shopxo.sql` 与 GoShop 结构后对照下文手工 SQL。若 `orders` 尚无 `payment_id` 等列，需先 `ALTER TABLE` 与当前模型一致。

## 可执行导入（推荐）

在同一 MySQL 实例上准备好**源库**（含 ShopXO `sxo_*` 表）与**目标库**（已通过 `bin/goshop` 或 `go run ./cmd/migrate` 建好 GoShop 表结构）后，一条命令完成与下文 SQL 等价的导入：

```bash
export GOSHOP_MYSQL_PASSWORD='你的密码'   # 或使用 -password
go run ./cmd/shopxo-import \
  -host 127.0.0.1 -user root \
  -from shopxo -to goshop \
  -wipe-target-tables \
  -reset-admin-password '新的管理员明文密码' -admin-id 1
```

说明：

- **环境变量**：本工具读取 **`GOSHOP_MYSQL_HOST` / `PORT` / `USER` / `PASSWORD`**（与 `cmd/shopxo-import` 一致）。**`scripts/migration_test.sh`** 为免与日常开发库混淆，使用单独前缀 **`GOSHOP_MIGRATION_MYSQL_*`**（默认口令 `goshop123` 对齐 `docker-compose.yml`）。
- **`-wipe-target-tables`** 会清空目标库中本迁移涉及的表（`users`、`orders` 等），**生产慎用**；测试库可先备份再执行。
- **`-table-prefix`**：若安装时不是默认 `sxo_`，传入实际前缀。
- **`-reset-admin-password`**：导入后管理员 `password` 列为占位，需 bcrypt 后才能登录管理端；也可用 SQL 自行 `UPDATE`。
- **管理员与 `roles`**：导入仅写入 `admins`，**不包含** ShopXO 侧与 GoShop `roles` 的一对一数据。若目标库尚无与 `admins.role_id` 匹配的角色行，管理端接口会 **403**。生产导入后请确保 **`roles` 表存在对应记录**（或先跑一次种子再按需覆盖业务表）。`migration_test.sh` 在导入后执行的 **`roles` 补全 SQL** 可作参考。
- **自测**：`scripts/migration_test.sh` 调用本工具（需本地 MySQL；测试期会临时改写 `config.yaml` 中 **db / redis.host**，退出时恢复）。

验算与调试用：**`-dry-run`**（只报 SQL 体积）、**`-print-sql`**（打印完整脚本到 stdout）。

手工复制本文 SQL 到 `mysql` 客户端时，请先 **`SET NAMES utf8mb4`** 或配置 **`default-character-set=utf8mb4`**（见上文「MySQL 客户端字符集」）。

---

## 第一步：数据库迁移

### 1.1 用户（`sxo_user`）

```sql
INSERT INTO goshop.users (id, username, password, nickname, phone, avatar, points, locking_integral, status, created_at, updated_at)
SELECT
  id,
  username,
  '',
  IFNULL(NULLIF(TRIM(nickname), ''), username),
  mobile,
  avatar,
  integral,
  locking_integral,
  CASE WHEN status = 0 THEN 1 ELSE 0 END,
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo.sxo_user
WHERE IFNULL(is_delete_time, 0) = 0;
```

### 1.2 商品分类（`sxo_goods_category`）

```sql
INSERT INTO goshop.categories (id, parent_id, name, icon, sort, status, created_at, updated_at)
SELECT
  id, pid, name, IFNULL(icon, ''),
  IFNULL(sort, 0),
  IFNULL(is_enable, 1),
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo.sxo_goods_category;
```

### 1.3 商品（`sxo_goods` + 分类关联 + 相册）

**说明**：`sort` 对应 ShopXO 的 `sort_level`；`give_integral`、`access_count` 一并迁入（与 `internal/model/goods.go` 一致）。

**相册 JSON（MySQL 8）**：无相册行时退化为仅含封面的单元素数组。

```sql
INSERT INTO goshop.goods (
  id, title, subtitle, category_id, brand_id, main_image, images, detail,
  status, sort, sales_count, access_count, give_integral, created_at, updated_at
)
SELECT
  g.id,
  g.title,
  IFNULL(g.simple_desc, ''),
  IFNULL(cj.category_id, 0),
  IFNULL(g.brand_id, 0),
  IFNULL(g.images, ''),
  COALESCE(
    (SELECT CAST(CONCAT('[', GROUP_CONCAT(JSON_QUOTE(p.images) ORDER BY IFNULL(p.sort, 0), p.id SEPARATOR ','), ']') AS JSON)
     FROM shopxo.sxo_goods_photo p
     WHERE p.goods_id = g.id AND IFNULL(p.is_show, 1) = 1),
    CASE WHEN IFNULL(g.images, '') <> '' THEN JSON_ARRAY(g.images) ELSE JSON_ARRAY() END
  ),
  IFNULL(g.content_web, ''),
  IFNULL(g.is_shelves, 0),
  IFNULL(g.sort_level, 0),
  IFNULL(g.sales_count, 0),
  IFNULL(g.access_count, 0),
  IFNULL(g.give_integral, 0),
  FROM_UNIXTIME(IF(g.add_time = 0, UNIX_TIMESTAMP(), g.add_time)),
  FROM_UNIXTIME(IF(IFNULL(g.upd_time, 0) = 0, g.add_time, g.upd_time))
FROM shopxo.sxo_goods g
LEFT JOIN (
  SELECT goods_id, MIN(category_id) AS category_id
  FROM shopxo.sxo_goods_category_join
  GROUP BY goods_id
) cj ON cj.goods_id = g.id
WHERE IFNULL(g.is_delete_time, 0) = 0;
```

### 1.4 商品 SKU（`sxo_goods_spec_base` + `sxo_goods_spec_value`）

**保留 ShopXO 的 `sxo_goods_spec_base.id`** 作为 `goshop.goods_skus.id`，便于与订单行、历史数据对齐。

`name`：将该规格下所有 `sxo_goods_spec_value.value` 用 ` / ` 拼接（与 ShopXO 展示习惯接近）。

```sql
INSERT INTO goshop.goods_skus (
  id, goods_id, name, price, stock, image, specs, coding, status, created_at, updated_at
)
SELECT
  b.id,
  b.goods_id,
  IFNULL(NULLIF(GROUP_CONCAT(DISTINCT v.value ORDER BY v.id SEPARATOR ' / '), ''), '默认'),
  ROUND(b.price * 100),
  IFNULL(b.inventory, 0),
  '',
  NULL,
  IFNULL(b.coding, ''),
  1,
  FROM_UNIXTIME(IF(b.add_time = 0, UNIX_TIMESTAMP(), b.add_time)),
  FROM_UNIXTIME(IF(b.add_time = 0, UNIX_TIMESTAMP(), b.add_time))
FROM shopxo.sxo_goods_spec_base b
LEFT JOIN shopxo.sxo_goods_spec_value v
  ON v.goods_spec_base_id = b.id AND v.goods_id = b.goods_id
GROUP BY b.id, b.goods_id, b.price, b.inventory, b.coding, b.add_time;
```

**无多规格商品**：ShopXO 可能没有 `sxo_goods_spec_base` 行，但订单仍按「单 SKU」计价。为 **`sku_id` 可解析**，插入占位 SKU（**人工约定 ID**：`100000000 + goods_id`，与下文订单明细一致；请确保不与现有 `spec_base.id` 冲突，一般官方演示数据 spec id 远小于 1e8）。

```sql
INSERT INTO goshop.goods_skus (
  id, goods_id, name, price, stock, image, specs, coding, status, created_at, updated_at
)
SELECT
  100000000 + g.id,
  g.id,
  '默认',
  ROUND(g.min_price * 100),
  IFNULL(g.inventory, 0),
  '',
  NULL,
  IFNULL(g.coding, ''),
  1,
  FROM_UNIXTIME(IF(g.add_time = 0, UNIX_TIMESTAMP(), g.add_time)),
  FROM_UNIXTIME(IF(IFNULL(g.upd_time, 0) = 0, g.add_time, g.upd_time))
FROM shopxo.sxo_goods g
WHERE IFNULL(g.is_delete_time, 0) = 0
  AND NOT EXISTS (
    SELECT 1 FROM shopxo.sxo_goods_spec_base b WHERE b.goods_id = g.id
  );
```

### 1.5 订单（`sxo_order` + `sxo_order_address`）

**地址快照**：拼成 JSON 写入 `orders.address`（字段名按常见约定；若 GoShop 前端有固定 schema，请与其对齐）。

**可选**：若 `pay_status = 2`（已退款）且业务希望落入 GoShop「已退款」，可扩展下面 `CASE`（示例：在 `status = 4` 已完成且已退款时映射为 `5`，需按你方财务规则调整）。

```sql
INSERT INTO goshop.orders (
  id, order_no, user_id, total_amount, pay_amount, status, remark, address,
  payment_id, order_model,
  created_at, updated_at, paid_at, shipped_at, completed_at
)
SELECT
  o.id,
  o.order_no,
  o.user_id,
  ROUND(o.total_price * 100),
  ROUND(o.pay_price * 100),
  CASE
    WHEN IFNULL(o.pay_status, 0) = 2 THEN 5
    ELSE CASE o.status
      WHEN 0 THEN 0
      WHEN 1 THEN 0
      WHEN 2 THEN 1
      WHEN 3 THEN 2
      WHEN 4 THEN 3
      WHEN 5 THEN 4
      WHEN 6 THEN 4
      ELSE 0
    END
  END,
  IFNULL(o.user_note, ''),
  IFNULL(
    (SELECT JSON_OBJECT(
      'name', oa.name,
      'phone', oa.tel,
      'province', oa.province_name,
      'city', oa.city_name,
      'district', oa.county_name,
      'detail', oa.address
    )
    FROM shopxo.sxo_order_address oa
    WHERE oa.order_id = o.id
    ORDER BY oa.id ASC
    LIMIT 1),
    ''
  ),
  IFNULL(o.payment_id, 0),
  IFNULL(o.order_model, 0),
  FROM_UNIXTIME(IF(o.add_time = 0, UNIX_TIMESTAMP(), o.add_time)),
  FROM_UNIXTIME(IF(IFNULL(o.upd_time, 0) = 0, o.add_time, o.upd_time)),
  IF(o.pay_time > 0, FROM_UNIXTIME(o.pay_time), NULL),
  IF(o.delivery_time > 0, FROM_UNIXTIME(o.delivery_time), NULL),
  IF(o.collect_time > 0, FROM_UNIXTIME(o.collect_time), NULL)
FROM shopxo.sxo_order o
WHERE IFNULL(o.is_delete_time, 0) = 0;
```

> **注意**：`pay_status = 2`（已退款）统一映射为 GoShop `status = 5`（已退款）。若存在「部分退款」等复杂状态，请结合 `sxo_order_aftersale` 与业务规则单独调整。

### 1.6 订单明细（`sxo_order_detail` → `order_items`）

**`sku_id` 解析**：

1. `spec_coding` 非空时，优先匹配 `sxo_goods_spec_base.coding`；
2. 否则取该商品最小 `spec_base.id`；
3. 仍无（无规格商品）→ `100000000 + goods_id`（与 1.4 占位 SKU 一致）。

```sql
INSERT INTO goshop.order_items (
  id, order_id, goods_id, sku_id, title, image, sku_name, price, quantity, created_at
)
SELECT
  d.id,
  d.order_id,
  d.goods_id,
  COALESCE(
    (SELECT b1.id FROM shopxo.sxo_goods_spec_base b1
     WHERE b1.goods_id = d.goods_id AND b1.coding = d.spec_coding AND IFNULL(d.spec_coding, '') <> ''
     LIMIT 1),
    (SELECT MIN(b2.id) FROM shopxo.sxo_goods_spec_base b2 WHERE b2.goods_id = d.goods_id),
    100000000 + d.goods_id
  ),
  IFNULL(d.title, ''),
  IFNULL(d.images, ''),
  IFNULL(NULLIF(TRIM(d.spec), ''), IFNULL(d.spec_desc, '')),
  ROUND(d.price * 100),
  IFNULL(d.buy_number, 0),
  FROM_UNIXTIME(IF(d.add_time = 0, UNIX_TIMESTAMP(), d.add_time))
FROM shopxo.sxo_order_detail d;
```

若部分历史行 `spec_coding` 与库内不一致导致匹配失败，会回退到 `MIN(id)` 或占位 id；迁移后请 **抽样核对** 订单明细价格与规格名。

### 1.7 收货地址（`sxo_user_address`）

v6.8 中 **`province` / `city` / `county` 为 `sxo_region.id`**，不是文字；**无** `province_name`、**无** `is_delete_time`。名称请 **`LEFT JOIN sxo_region`**（与 `sxo_order_address` 存 `*_name` 不同）。

```sql
INSERT INTO goshop.addresses (id, user_id, name, phone, province, city, district, detail, is_default, created_at, updated_at)
SELECT
  ua.id,
  ua.user_id,
  ua.name,
  ua.tel,
  IFNULL(rp.name, ''),
  IFNULL(rc.name, ''),
  IFNULL(rd.name, ''),
  IFNULL(ua.address, ''),
  IFNULL(ua.is_default, 0),
  FROM_UNIXTIME(IF(ua.add_time = 0, UNIX_TIMESTAMP(), ua.add_time)),
  FROM_UNIXTIME(IF(IFNULL(ua.upd_time, 0) = 0, ua.add_time, ua.upd_time))
FROM shopxo.sxo_user_address ua
LEFT JOIN shopxo.sxo_region rp ON rp.id = ua.province
LEFT JOIN shopxo.sxo_region rc ON rc.id = ua.city
LEFT JOIN shopxo.sxo_region rd ON rd.id = ua.county;
```

### 1.8 品牌 / 文章 / 管理员

```sql
-- 品牌（sxo_brand）
INSERT INTO goshop.brands (id, name, logo, sort, status, created_at, updated_at)
SELECT
  id, name, IFNULL(logo, ''), IFNULL(sort, 0), IFNULL(is_enable, 1),
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo.sxo_brand;

-- 文章（sxo_article 无 sort 列，用 0；is_enable → status：1 发布 / 0 草稿）
INSERT INTO goshop.articles (
  id, title, content, category_id, cover, access_count, sort, status, created_at, updated_at
)
SELECT
  id,
  title,
  IFNULL(content, ''),
  IFNULL(article_category_id, 0),
  IFNULL(cover, ''),
  IFNULL(access_count, 0),
  0,
  IF(IFNULL(is_enable, 1) = 1, 1, 0),
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo.sxo_article;

-- 管理员：password 须后续 bcrypt 更新；status 反转
INSERT INTO goshop.admins (id, username, password, nickname, role_id, status, created_at, updated_at)
SELECT
  id,
  username,
  '',
  IFNULL(NULLIF(TRIM(mobile), ''), username),
  IFNULL(role_id, 1),
  CASE WHEN status = 0 THEN 1 ELSE 0 END,
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo.sxo_admin;
```

---

## 第二步：密码兼容

ShopXO 用户：`pwd` + `salt`（md5）。GoShop：`password`（bcrypt）。**须**方案 A 重置或方案 B 自改代码（见旧版说明，此处不重复）。

```sql
-- 管理员临时密码示例（bcrypt of "admin123"，仅演示）
UPDATE goshop.admins SET password = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' WHERE id = 1;
```

### 方案 B（未内置）：模型 + 登录补丁（示意）

若需 md5 渐进迁移，仍可按此前文档在 `users` 表增加 `salt`、`old_pwd` 并在 `Login` 中 bcrypt 失败后校验 md5；**当前仓库未包含此逻辑**。

---

## 第三步：文件迁移

```bash
cp -r /path/to/shopxo/public/static/upload/* /path/to/goshop/uploads/
```

```sql
UPDATE goshop.goods SET main_image = REPLACE(main_image, '/static/upload/', '/uploads/') WHERE main_image LIKE '%/static/upload/%';
UPDATE goshop.goods SET images = REPLACE(images, '/static/upload/', '/uploads/') WHERE images LIKE '%/static/upload/%';
UPDATE goshop.articles SET cover = REPLACE(cover, '/static/upload/', '/uploads/') WHERE cover LIKE '%/static/upload/%';
UPDATE goshop.brands SET logo = REPLACE(logo, '/static/upload/', '/uploads/') WHERE logo LIKE '%/static/upload/%';
UPDATE goshop.order_items SET image = REPLACE(image, '/static/upload/', '/uploads/') WHERE image LIKE '%/static/upload/%';
-- 若 specs / JSON 内嵌路径，需单独脚本处理
```

---

## 第四步：uni-app / 前端

- **shopxo-uniapp**：改 `App.vue` → `globalData.data.request_url` 为 GoShop 站点根（尾斜杠），详见 `docs/uniapp-guide.md` **方案 B**。
- **PC/管理端**：使用 GoShop 自带前端，不迁 PHP 模板。

---

## 建议迁移顺序

1. `users` → `categories` → `brands`
2. `goods`（含分类子查询、相册）
3. `goods_skus`（先 `spec_base`，再「无规格占位」）
4. `addresses`
5. `orders`（含 `order_address` JSON）
6. `order_items`
7. `articles`
8. `admins`
9. 文件与路径替换、重置密码、`AUTO_INCREMENT`（如下）

### 迁移后自增

对曾手工指定 `id` 插入的表，将 `AUTO_INCREMENT` 设为 **`MAX(id) + 1`**，避免新数据主键冲突。不同 MySQL 版本对 `ALTER TABLE … = (子查询)` 支持不一，建议在客户端对每张表执行：

```sql
SELECT MAX(id) + 1 AS next_ai FROM goshop.users;
-- 然后将 users 表的 AUTO_INCREMENT 设为查询结果（例如在 Workbench / phpMyAdmin 中改表选项）
```

至少检查：`users`、`categories`、`goods`、`goods_skus`、`orders`、`order_items`、`addresses`、`articles`、`admins`、`brands`。

---

## 迁移检查清单

- [ ] 源库版本与 `shopxo.sql` / 真实 `SHOW CREATE TABLE` 一致
- [ ] 用户、分类、商品、SKU、订单主从、地址条数抽样一致
- [ ] 金额（元→分）、订单与明细行 `price * quantity` 与主表核对
- [ ] 无规格商品的 `100000000+goods_id` SKU 与订单明细能对应
- [ ] 管理员可登录；用户密码策略已落地
- [ ] 图片路径、相册 JSON（`GROUP_CONCAT`+`JSON_QUOTE` 方案）已验证
- [ ] uni-app / 收银台（若使用）与 `internal/compat/shopxo` 行为已联调

---

## 不迁移的内容（示例）

- `sxo_config`、`sxo_power`、`sxo_region` 等由 GoShop 初始化或另行对接
- PHP 插件、主题模板
- 售后、优惠券、营销等扩展表：按需另写 SQL，不在本文范围

---

## 常见问题

**Q: v5.x 能迁吗？**  
A: 表名/字段可能不同，请对照你库 DDL 改 SQL。

**Q: 多分类商品只迁了一个 category_id？**  
A: 是的；多对多需改 GoShop 模型或选定业务规则后再迁。

**Q: 占位 SKU id `1e8+goods_id` 会冲突吗？**  
A: 在 spec_base id 远小于 1e8 的常规数据下安全；若你方 spec id 已超界，请改用更大前缀并在 1.4 / 1.6 同步修改。

**Q: 旧文档里的 `address_data`、`category_id`、`photo` 列？**  
A: 均 **不是** v6.8 `sxo_goods` / `sxo_order` 的标准列，已按 `/tmp/shopxo/config/shopxo.sql` 更正。
