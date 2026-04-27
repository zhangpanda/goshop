# ShopXO 迁移到 GoShop 指南

## 适用版本

- **ShopXO v6.x**（本文档基于 v6.8.0 编写和验证）
- 表前缀默认 `sxo_`（安装时可自定义，下文以 `sxo_` 为例）
- v5.x 及更早版本表结构差异较大，需自行调整 SQL

> GoShop 在数据模型上与 ShopXO v6.8.0 常见库表做过对照，便于迁移；应用代码为独立实现。

## 迁移概览

| 步骤 | 内容 | 难度 |
|------|------|------|
| 1 | 数据库迁移（SQL 脚本） | 中 |
| 2 | 密码兼容处理 | 低 |
| 3 | 上传文件迁移 | 低 |
| 4 | uni-app 前端切换 | 低（改一行配置） |

## 核心差异对照

| 差异项 | ShopXO v6.8.0 | GoShop |
|--------|---------------|--------|
| 表前缀 | `sxo_`（可自定义） | 无前缀 |
| 金额单位 | 元（decimal） | 分（int64） |
| 时间格式 | Unix 时间戳（int） | datetime |
| 用户密码 | `md5(salt + pwd)` | bcrypt |
| 用户状态 | 0=正常, 1=禁言, 2=禁登 | 0=禁用, 1=正常 |
| 订单状态 | 0待确认,1待支付,2待发货,3待收货,4完成,5取消,6关闭 | 0待付款,1待发货,2待收货,3完成,4取消,5退款 |
| 管理员密码字段 | `login_pwd` + `login_salt` | `password`(bcrypt) |

## 前置条件

- GoShop 已部署并首次启动过（自动建表）
- 能访问 ShopXO 的 MySQL 数据库
- **先在测试环境跑一遍，确认无误再操作生产数据**

## 第一步：数据库迁移

以下 SQL 假设 ShopXO 库名 `shopxo`、表前缀 `sxo_`，GoShop 库名 `goshop`。

> 如果你的 ShopXO 表前缀不是 `sxo_`，全局替换即可。

### 1.1 用户

```sql
INSERT INTO goshop.users (id, username, password, nickname, phone, avatar, points, locking_integral, status, created_at, updated_at)
SELECT
  id, username, '',
  IF(nickname = '', username, nickname),
  mobile, avatar,
  integral, locking_integral,
  CASE WHEN status = 0 THEN 1 ELSE 0 END,  -- ShopXO 0=正常 → GoShop 1=正常
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_user
WHERE is_delete_time = 0;
```

### 1.2 商品分类

```sql
INSERT INTO goshop.categories (id, parent_id, name, icon, sort, status, created_at, updated_at)
SELECT id, pid, name, IFNULL(icon, ''), IFNULL(sort, 0),
  IFNULL(is_enable, 1),
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_goods_category;
```

### 1.3 商品

```sql
-- 注意：金额 × 100 从元转分
INSERT INTO goshop.goods (id, title, subtitle, category_id, brand_id, main_image, images, detail, status, sort, sales_count, created_at, updated_at)
SELECT id, title, simple_desc, IFNULL(category_id, 0), IFNULL(brand_id, 0),
  images, IFNULL(photo, ''), IFNULL(content_web, ''),
  IFNULL(is_shelves, 0), IFNULL(sort, 0), IFNULL(sales_count, 0),
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_goods;
```

### 1.4 订单

```sql
-- 金额 × 100；状态映射：ShopXO(0待确认→1待支付→2待发货→3待收货→4完成→5取消→6关闭)
-- GoShop(0待付款→1待发货→2待收货→3完成→4取消)
INSERT INTO goshop.orders (id, order_no, user_id, total_amount, pay_amount, status, remark, address, created_at, updated_at, paid_at, shipped_at, completed_at)
SELECT id, order_no, user_id,
  ROUND(total_price * 100), ROUND(pay_price * 100),
  CASE status
    WHEN 0 THEN 0  -- 待确认 → 待付款
    WHEN 1 THEN 0  -- 待支付 → 待付款
    WHEN 2 THEN 1  -- 待发货
    WHEN 3 THEN 2  -- 待收货
    WHEN 4 THEN 3  -- 已完成
    WHEN 5 THEN 4  -- 已取消
    WHEN 6 THEN 4  -- 已关闭 → 已取消
    ELSE 0
  END,
  user_note,
  IFNULL(address_data, ''),
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time)),
  IF(pay_time > 0, FROM_UNIXTIME(pay_time), NULL),
  IF(delivery_time > 0, FROM_UNIXTIME(delivery_time), NULL),
  IF(collect_time > 0, FROM_UNIXTIME(collect_time), NULL)
FROM shopxo.sxo_order;
```

### 1.5 收货地址

```sql
INSERT INTO goshop.addresses (id, user_id, name, phone, province, city, district, detail, is_default, created_at, updated_at)
SELECT id, user_id, name, tel,
  IFNULL(province_name, ''), IFNULL(city_name, ''), IFNULL(county_name, ''),
  IFNULL(address, ''), IFNULL(is_default, 0),
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_user_address
WHERE is_delete_time = 0;
```

### 1.6 品牌 / 文章 / 管理员

```sql
-- 品牌
INSERT INTO goshop.brands (id, name, logo, sort, status, created_at, updated_at)
SELECT id, name, IFNULL(logo, ''), IFNULL(sort, 0), IFNULL(is_enable, 1),
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_brand;

-- 文章
INSERT INTO goshop.articles (id, title, content, category_id, cover, status, sort, created_at, updated_at)
SELECT id, title, IFNULL(content, ''), IFNULL(article_category_id, 0),
  IFNULL(image, ''), IFNULL(is_enable, 1), IFNULL(sort, 0),
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_article;

-- 管理员（密码留空，见第二步）
INSERT INTO goshop.admins (id, username, password, nickname, role_id, status, created_at, updated_at)
SELECT id, username, '', IFNULL(mobile, username), IFNULL(role_id, 1), status,
  FROM_UNIXTIME(IF(add_time=0, UNIX_TIMESTAMP(), add_time)),
  FROM_UNIXTIME(IF(upd_time=0, add_time, upd_time))
FROM shopxo.sxo_admin;
```

## 第二步：密码兼容

ShopXO 密码：`md5(salt + 明文密码)`，存在 `pwd` 和 `salt` 两个字段。
GoShop 密码：`bcrypt(明文密码)`，存在 `password` 一个字段。

md5 不可逆，无法直接转换。两种方案：

### 方案 A：强制重置（推荐，零代码改动）

迁移后通知用户通过「忘记密码」重置。管理员可在后台手动设置密码。

```sql
-- 给管理员设一个临时密码（bcrypt hash of "admin123"）
UPDATE goshop.admins SET password = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy' WHERE id = 1;
```

### 方案 B：兼容登录（渐进迁移，需改代码）

在 User model 加两个临时字段，登录时先试 bcrypt，失败再试 md5，成功后自动升级：

```go
// internal/model/user.go — 临时增加
Salt   string `json:"-" gorm:"size:32;comment:ShopXO迁移salt"`
OldPwd string `json:"-" gorm:"column:old_pwd;size:32;comment:ShopXO迁移md5密码"`
```

```sql
-- 把 ShopXO 的 salt 和 pwd 写入临时字段
ALTER TABLE goshop.users ADD COLUMN salt VARCHAR(32) DEFAULT '' AFTER password;
ALTER TABLE goshop.users ADD COLUMN old_pwd VARCHAR(32) DEFAULT '' AFTER salt;

UPDATE goshop.users u
JOIN shopxo.sxo_user su ON u.id = su.id
SET u.salt = su.salt, u.old_pwd = su.pwd;
```

在登录逻辑中 bcrypt 校验失败后加兼容：

```go
// internal/service/user.go Login 函数中，bcrypt 失败后：
if user.Salt != "" && user.OldPwd != "" {
    md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(user.Salt+req.Password)))
    if md5Hash == user.OldPwd {
        // 自动升级为 bcrypt
        newHash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        global.DB.Model(&user).Updates(map[string]interface{}{
            "password": string(newHash), "salt": "", "old_pwd": "",
        })
        // 继续正常登录...
    }
}
```

所有用户登录一次后 salt/old_pwd 自动清空，之后可删除这两个字段。

## 第三步：文件迁移

ShopXO 上传文件在 `public/static/upload/` 或 `public/static/plugins/`。

```bash
cp -r /path/to/shopxo/public/static/upload/* /path/to/goshop/uploads/
```

批量替换数据库中的图片路径：

```sql
UPDATE goshop.goods SET main_image = REPLACE(main_image, '/static/upload/', '/uploads/') WHERE main_image LIKE '%/static/upload/%';
UPDATE goshop.goods SET images = REPLACE(images, '/static/upload/', '/uploads/') WHERE images LIKE '%/static/upload/%';
UPDATE goshop.articles SET cover = REPLACE(cover, '/static/upload/', '/uploads/') WHERE cover LIKE '%/static/upload/%';
UPDATE goshop.brands SET logo = REPLACE(logo, '/static/upload/', '/uploads/') WHERE logo LIKE '%/static/upload/%';
```

## 第四步：前端切换

### uni-app 手机端（改一行）

GoShop 内置 `/api.php` 兼容层，当前实现 **82** 个 `s=` 动作，便于与常见 shopxo-uniapp 对接（详见 `internal/handler/shopxo_compat.go` 与 `HANDOVER.md`；**非** ShopXO 官方组件）：

```javascript
// shopxo-uniapp/common/config.js
request_url: 'https://your-goshop-domain.com'
```

### PC 前台 & 管理后台

GoShop 自带全新前端，无需迁移 ShopXO 的 PHP 模板。

## 迁移检查清单

- [ ] 确认 ShopXO 版本为 v6.x
- [ ] 确认表前缀（默认 `sxo_`，自定义的需替换 SQL 中的前缀）
- [ ] 用户数量一致（排除已删除用户）
- [ ] 商品数量一致，图片能正常显示
- [ ] 订单数量一致，金额正确（元→分，检查几条对比）
- [ ] 订单状态映射正确
- [ ] 管理员能登录后台
- [ ] 上传文件已复制，路径已替换
- [ ] uni-app 能正常访问（如使用）
- [ ] 测试下单、支付、发货全流程

## 不迁移的内容

以下 ShopXO 功能/数据不需要迁移（GoShop 会自动初始化）：

- **系统配置**（`sxo_config`）— GoShop 首次启动自动生成 77 项配置
- **权限节点**（`sxo_power`）— GoShop 自动初始化
- **省市区**（`sxo_region`）— GoShop 自动初始化
- **PHP 插件** — 不兼容，GoShop 有独立的功能实现
- **PHP 主题模板** — GoShop 使用 React/Next.js 前端

## 常见问题

**Q: v5.x 能迁移吗？**
A: 表结构有差异（字段名、状态码），需要自行对照调整 SQL。核心思路一样。

**Q: 迁移后 ShopXO 还能用吗？**
A: 可以。迁移是读取 ShopXO 数据写入 GoShop，不修改原库。可以并行运行一段时间。

**Q: 订单金额有小数怎么办？**
A: `ROUND(price * 100)` 会四舍五入到分。如果原始数据有 0.001 元这种精度，会丢失。

**Q: 自定义表前缀怎么处理？**
A: 把 SQL 中所有 `sxo_` 替换为你的实际前缀即可。
