-- 已弃用维护：权威脚本已嵌入仓库，请使用：
--   go run ./cmd/shopxo-import -from <ShopXO库> -to <GoShop库> [可选 -wipe-target-tables]
-- 源码与副本：internal/shopxomigrate/data.sql（与 docs/migration-from-shopxo.md 同步）
--
-- 以下为历史占位名示例（仅当仍需手工 sed 时）：
--   源 shopxo_mig_src、目标 goshop_mig_dst

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ========== 1.1 用户 ==========
INSERT INTO goshop_mig_dst.users (id, username, password, nickname, phone, avatar, points, locking_integral, status, created_at, updated_at)
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
FROM shopxo_mig_src.sxo_user
WHERE IFNULL(is_delete_time, 0) = 0;

-- ========== 1.2 分类 ==========
INSERT INTO goshop_mig_dst.categories (id, parent_id, name, icon, sort, status, created_at, updated_at)
SELECT
  id, pid, name, IFNULL(icon, ''),
  IFNULL(sort, 0),
  IFNULL(is_enable, 1),
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo_mig_src.sxo_goods_category;

-- ========== 1.3 商品 ==========
INSERT INTO goshop_mig_dst.goods (
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
     FROM shopxo_mig_src.sxo_goods_photo p
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
FROM shopxo_mig_src.sxo_goods g
LEFT JOIN (
  SELECT goods_id, MIN(category_id) AS category_id
  FROM shopxo_mig_src.sxo_goods_category_join
  GROUP BY goods_id
) cj ON cj.goods_id = g.id
WHERE IFNULL(g.is_delete_time, 0) = 0;

-- ========== 1.4 SKU（spec_base） ==========
INSERT INTO goshop_mig_dst.goods_skus (
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
FROM shopxo_mig_src.sxo_goods_spec_base b
LEFT JOIN shopxo_mig_src.sxo_goods_spec_value v
  ON v.goods_spec_base_id = b.id AND v.goods_id = b.goods_id
GROUP BY b.id, b.goods_id, b.price, b.inventory, b.coding, b.add_time;

-- ========== 1.4b 无规格占位 SKU ==========
INSERT INTO goshop_mig_dst.goods_skus (
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
FROM shopxo_mig_src.sxo_goods g
WHERE IFNULL(g.is_delete_time, 0) = 0
  AND NOT EXISTS (
    SELECT 1 FROM shopxo_mig_src.sxo_goods_spec_base b WHERE b.goods_id = g.id
  );

-- ========== 1.5 订单 ==========
INSERT INTO goshop_mig_dst.orders (
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
    FROM shopxo_mig_src.sxo_order_address oa
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
FROM shopxo_mig_src.sxo_order o
WHERE IFNULL(o.is_delete_time, 0) = 0;

-- ========== 1.6 订单明细 ==========
INSERT INTO goshop_mig_dst.order_items (
  id, order_id, goods_id, sku_id, title, image, sku_name, price, quantity, created_at
)
SELECT
  d.id,
  d.order_id,
  d.goods_id,
  COALESCE(
    (SELECT b1.id FROM shopxo_mig_src.sxo_goods_spec_base b1
     WHERE b1.goods_id = d.goods_id AND b1.coding = d.spec_coding AND IFNULL(d.spec_coding, '') <> ''
     LIMIT 1),
    (SELECT MIN(b2.id) FROM shopxo_mig_src.sxo_goods_spec_base b2 WHERE b2.goods_id = d.goods_id),
    100000000 + d.goods_id
  ),
  IFNULL(d.title, ''),
  IFNULL(d.images, ''),
  IFNULL(NULLIF(TRIM(d.spec), ''), IFNULL(d.spec_desc, '')),
  ROUND(d.price * 100),
  IFNULL(d.buy_number, 0),
  FROM_UNIXTIME(IF(d.add_time = 0, UNIX_TIMESTAMP(), d.add_time))
FROM shopxo_mig_src.sxo_order_detail d;

-- ========== 1.7 收货地址（v6.8：省市区为 region id，JOIN sxo_region） ==========
INSERT INTO goshop_mig_dst.addresses (id, user_id, name, phone, province, city, district, detail, is_default, created_at, updated_at)
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
FROM shopxo_mig_src.sxo_user_address ua
LEFT JOIN shopxo_mig_src.sxo_region rp ON rp.id = ua.province
LEFT JOIN shopxo_mig_src.sxo_region rc ON rc.id = ua.city
LEFT JOIN shopxo_mig_src.sxo_region rd ON rd.id = ua.county;

-- ========== 1.8 品牌 / 文章 / 管理员 ==========
INSERT INTO goshop_mig_dst.brands (id, name, logo, sort, status, created_at, updated_at)
SELECT
  id, name, IFNULL(logo, ''), IFNULL(sort, 0), IFNULL(is_enable, 1),
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo_mig_src.sxo_brand;

INSERT INTO goshop_mig_dst.articles (
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
FROM shopxo_mig_src.sxo_article;

INSERT INTO goshop_mig_dst.admins (id, username, password, nickname, role_id, status, created_at, updated_at)
SELECT
  id,
  username,
  '',
  IFNULL(NULLIF(TRIM(mobile), ''), username),
  IFNULL(role_id, 1),
  CASE WHEN status = 0 THEN 1 ELSE 0 END,
  FROM_UNIXTIME(IF(add_time = 0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(IFNULL(upd_time, 0) = 0, add_time, upd_time))
FROM shopxo_mig_src.sxo_admin;

SET FOREIGN_KEY_CHECKS = 1;
