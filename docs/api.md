# GoShop API 文档

本文随开发迭代更新，**非**对外固定契约；生产接入请以路由与 handler 实现为准。

基础地址：`http://localhost:8080`

响应格式：`{"code": 0, "msg": "success", "data": {...}}`，code=0 表示成功。

价格单位：所有价格以**分**为单位（int64），前端显示时除以 100。

## 认证

需要登录的接口在 Header 中传递：`Authorization: Bearer {token}`

---

## 公共接口（无需登录）

### 站点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/site-config | 站点配置（名称/Logo/SEO/公告等） |
| GET | /api/slides | 轮播图列表 |
| GET | /api/navigations?type=header | 导航菜单（header/footer） |
| GET | /api/links | 友情链接 |
| GET | /api/payments | 支付方式列表 |
| GET | /api/regions | 地区列表（?parent_id=0） |
| GET | /api/region/all | 全部地区（省市区三级） |
| GET | /api/express | 快递公司列表 |
| GET | /api/agreement | 协议内容 |
| GET | /api/quick-nav | 快捷导航 |

### 商品

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| GET | /api/categories | 分类树（含二级） | - |
| GET | /api/goods | 商品列表 | keyword, category_id, status, page, page_size |
| GET | /api/goods/:id | 商品详情 | - |
| GET | /api/goods/:id/reviews | 商品评价 | page, page_size |
| GET | /api/goods/:id/specs | 商品规格 | - |
| GET | /api/goods/:id/params | 商品参数 | - |
| GET | /api/goods/:id/photos | 商品相册 | - |
| GET | /api/goods/:id/stock | 库存查询 | - |
| GET | /api/brands | 品牌列表 | - |
| GET | /api/coupons | 可领优惠券列表 | - |
| GET | /api/promotions | 进行中的促销 | - |
| GET | /api/seckills | 进行中的秒杀活动 | - |
| GET | /api/group-buys | 进行中的拼团活动 | - |
| GET | /api/group-orders/:id | 拼团详情（团员列表） | - |
| GET | /api/search/hot | 热门搜索词 | - |
| GET | /api/search/prices | 价格筛选区间 | - |

### 文章

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/article-categories | 文章分类 |
| GET | /api/articles | 文章列表（?category_id=） |
| GET | /api/articles/:id | 文章详情 |

### 用户认证

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | /api/register | 注册 | username, password, nickname |
| POST | /api/login | 登录 | username, password |
| POST | /api/wx/login | 微信小程序登录 | code, nickname, avatar |
| POST | /api/forget-password | 忘记密码 | - |
| POST | /api/verify-code | 发送验证码 | phone/email, type |

---

## 用户接口（需登录）

### 用户信息

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/user/profile | 获取个人信息 |
| PUT | /api/user/password | 修改密码 |
| POST | /api/user/bind-mobile | 绑定手机 |
| POST | /api/user/logout | 退出登录 |

### 购物车

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | /api/cart | 加入购物车 | goods_id, sku_id, quantity |
| GET | /api/cart | 购物车列表 | - |
| PUT | /api/cart/:id | 修改数量 | quantity |
| DELETE | /api/cart | 批量删除 | ids (数组) |
| PUT | /api/cart/select-all | 全选/取消 | selected (bool) |

### 收货地址

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | /api/address | 新增地址 | name, phone, province, city, district, detail, is_default |
| GET | /api/address | 地址列表 | - |
| PUT | /api/address/:id | 修改地址 | 同上 |
| DELETE | /api/address/:id | 删除地址 | - |

### 订单

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | /api/orders | 创建订单 | address_id, cart_ids, user_coupon_id |
| GET | /api/orders | 订单列表 | status, page, page_size |
| GET | /api/orders/:id | 订单详情 | - |
| PUT | /api/orders/:id/cancel | 取消订单 | - |
| PUT | /api/orders/:id/receive | 确认收货 | - |
| GET | /api/orders/:id/shipment | 物流信息 | - |
| DELETE | /api/orders/:id | 删除订单 | - |

### 支付

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | /api/pay | 发起支付（微信JSAPI） | order_id, openid |
| POST | /api/pay/unified | 统一支付入口 | order_id, payment_key, openid, return_url |
| POST | /api/pay/refund | 申请退款 | order_id, reason |
| POST | /api/pay/notify | 微信支付回调（系统调用） | - |
| POST | /api/pay/alipay-notify | 支付宝支付回调（系统调用，含RSA2验签） | - |
| POST | /api/pay/log | 创建支付日志 | - |
| GET | /api/pay/log | 支付日志列表 | - |

> `payment_key` 可选值：`wechat_jsapi` / `wechat_h5` / `wechat_app` / `wechat_native` / `alipay_pc` / `alipay_h5` / `alipay_app` / `alipay_mini` / `wallet` / `offline`

### 秒杀 & 拼团

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| POST | /api/seckill/:item_id/buy | 秒杀抢购 | - |
| POST | /api/group/:id/open | 开团 | - |
| POST | /api/group/:id/join | 参团 | - |

> 拼团路径中两段均为 `…/group/:id/…`：`/open` 的 `id` 为**活动商品/拼团项** ID；`/join` 的 `id` 为**拼团单** ID。

### 售后

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/aftersale | 申请售后 |
| GET | /api/aftersale | 售后列表 |
| GET | /api/aftersale/:id | 售后详情 |
| PUT | /api/aftersale/:id/cancel | 取消售后 |
| PUT | /api/aftersale/:id/delivery | 退货发货 |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/coupons/:id/receive | 领取优惠券 |
| GET | /api/my/coupons | 我的优惠券（?status=0未使用/1已使用/2已过期） |
| POST | /api/favorites/:id | 收藏/取消收藏 |
| GET | /api/favorites | 收藏列表 |
| GET | /api/history | 浏览记录 |
| DELETE | /api/history | 清空浏览记录 |
| POST | /api/reviews | 发表评价 |
| POST | /api/points/sign | 签到 |
| GET | /api/points/log | 积分记录 |
| GET | /api/messages | 消息列表 |
| PUT | /api/messages/:id/read | 标记已读 |
| PUT | /api/messages/read-all | 全部已读 |
| POST | /api/upload | 文件上传（multipart/form-data, field: file） |

---

## 管理后台接口（/api/admin/*）

需要管理员 token，通过 `/api/admin/login` 获取。

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/captcha?key=xxx | 获取图片验证码 |
| POST | /api/admin/login | 管理员登录（username, password, captcha_key, captcha_code） |

### 仪表盘

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/dashboard | 统计数据（商品数/订单数/用户数/销售额/趋势） |

### 商品管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/goods | 创建商品（含 SKU） |
| PUT | /api/admin/goods/:id | 修改商品 |
| DELETE | /api/admin/goods/:id | 删除商品 |
| PUT | /api/admin/goods/:id/status | 上架/下架 |
| POST | /api/admin/categories | 创建分类 |
| PUT | /api/admin/categories/:id | 修改分类 |
| DELETE | /api/admin/categories/:id | 删除分类 |

### 订单管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/orders | 订单列表 |
| PUT | /api/admin/orders/:id/remark | 订单备注 |
| POST | /api/admin/orders/ship | 发货（order_id, express_company, express_no） |
| PUT | /api/admin/orders/:id/cancel | 取消订单 |

### 用户管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/users | 用户列表 |
| PUT | /api/admin/users/:id/status | 禁用/启用 |

### 内容管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/admin/articles | 创建文章 |
| PUT | /api/admin/articles/:id | 修改文章 |
| DELETE | /api/admin/articles/:id | 删除文章 |
| POST | /api/admin/slides | 创建轮播图 |
| PUT | /api/admin/slides/:id | 修改轮播图 |
| POST | /api/admin/navigations | 创建导航 |
| PUT | /api/admin/navigations/:id | 修改导航 |
| POST | /api/admin/brands | 创建品牌 |
| POST | /api/admin/coupons | 创建优惠券 |
| POST | /api/admin/promotions | 创建促销 |
| PUT | /api/admin/reviews/:id/reply | 回复评价 |

### 秒杀 & 拼团管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/seckills | 秒杀活动列表 |
| POST | /api/admin/seckills | 创建秒杀活动 |
| GET | /api/admin/group-buys | 拼团活动列表 |
| POST | /api/admin/group-buys | 创建拼团活动 |

### 系统配置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/config?group=base | 获取配置（base/site/seo/email/sms/order/app 等） |
| POST | /api/admin/config | 保存配置（key, value, group, desc） |

### 权限管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/admins | 管理员列表 |
| POST | /api/admin/admins | 创建管理员 |
| GET | /api/admin/roles | 角色列表 |
| POST | /api/admin/roles | 创建角色 |
| GET | /api/admin/powers | 权限树 |
| PUT | /api/admin/roles/:id/powers | 分配权限 |

---

## ShopXO 兼容接口

> **ShopXO** 为独立项目；本节描述 GoShop 提供的 **`/api.php` 请求形态**，便于与常见 shopxo-uniapp 对接，**不声称**与对方任意发行版逐接口逐字段一致。

通过 `/api.php?s=controller/action` 调用，token 通过 `?token=xxx` 传递。

当前实现注册 **82** 个 `s=` 动作（`routeMap`，详见 `internal/handler/shopxo_compat.go`）；**边界行为以源码与测试为准**。使用说明见 [uni-app 对接指南](uniapp-guide.md)。

常用接口映射：

| ShopXO 接口 | GoShop 接口 | 说明 |
|---|---|---|
| api.php?s=index/index | /api/goods + /api/slides | 首页 |
| api.php?s=goods/detail&goods_id=1 | /api/goods/1 | 商品详情 |
| api.php?s=goods/category | /api/categories | 分类 |
| api.php?s=user/login | /api/login | 登录（支持 accounts/pwd 字段） |
| api.php?s=cart/index | /api/cart | 购物车 |
| api.php?s=order/index | /api/orders | 订单列表 |
| api.php?s=search/index&keywords=xxx | /api/goods?keyword=xxx | 搜索 |
| api.php?s=cashier/paydata | `/api.php?s=cashier/paydata` | 微信小程序收银台：参数 `authcode`（wx.login）、`order_no`（order/pay 返回的 PayLog `order_no`）；需用户已绑定与订单一致的微信 openid |
