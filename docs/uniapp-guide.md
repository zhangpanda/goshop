# GoShop uni-app 前端对接指南

> **ShopXO** / **shopxo-uniapp** 为独立开源项目；本文仅说明如何与 GoShop 后端对接，**不代表**官方合作或背书。

## 概述

GoShop 支持两种 uni-app 对接方式，**不要混用**：

| 方案 | 说明 | 本文位置 |
|------|------|----------|
| **方案 A：GoShop 原生 REST** | 请求形如 `https://域名/api/...`，`Authorization: Bearer <token>`，与 Go 侧路由一致。 | 第 1～3 节 |
| **方案 B：shopxo-uniapp + `api.php`** | 请求形如 `https://域名/api.php?s=模块/方法`，沿用 ShopXO 系前端的拼 URL 方式；依赖 `internal/compat/shopxo` 兼容层。 | 第 4 节 |

**方案 A 的 `utils/request.js` 与方案 B 的官方前端不是同一套**：不能只改 `BASE_URL` 就指望官方 shopxo-uniapp 走 REST；官方项目默认走 **方案 B**。

---

## 1. API 适配层（方案 A：原生 REST）

将以下文件放到 uni-app 项目的 `utils/` 目录下：

### utils/request.js

```javascript
/**
 * GoShop 原生 REST 前缀（含 `/api`），须与部署路径一致。
 * 与 `uploadFile` 等共用，避免误用 `request.BASE_URL`（默认导出函数上不存在该属性）。
 */
export const API_BASE = 'http://localhost:8080/api'  // 开发环境，上线改为你的域名

/**
 * @param {{ url: string, method?: string, data?: object, header?: Record<string, string> }} options
 * @returns {Promise<unknown>}
 */
const request = (options) => {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token')
    uni.request({
      url: API_BASE + options.url,
      method: options.method || 'GET',
      data: options.data,
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : '',
        ...options.header
      },
      success: (res) => {
        if (res.data.code === 0) {
          resolve(res.data.data)
        } else if (res.statusCode === 401) {
          uni.removeStorageSync('token')
          uni.navigateTo({ url: '/pages/login/login' })
          reject(res.data)
        } else {
          uni.showToast({ title: res.data.msg, icon: 'none' })
          reject(res.data)
        }
      },
      fail: reject
    })
  })
}

export default request
```

### utils/api.js

```javascript
import request, { API_BASE } from './request'

// ========== 用户 ==========
export const wxLogin = (data) => request({ url: '/wx/login', method: 'POST', data })
export const login = (data) => request({ url: '/login', method: 'POST', data })
export const register = (data) => request({ url: '/register', method: 'POST', data })
export const getUserProfile = () => request({ url: '/user/profile' })

// ========== 商品 ==========
export const getCategories = () => request({ url: '/categories' })
export const getGoodsList = (params) => request({ url: '/goods', data: params })
export const getGoodsDetail = (id) => request({ url: `/goods/${id}` })
export const getGoodsReviews = (id, params) => request({ url: `/goods/${id}/reviews`, data: params })

// ========== 购物车 ==========
export const addCart = (data) => request({ url: '/cart', method: 'POST', data })
export const getCartList = () => request({ url: '/cart' })
export const updateCart = (id, data) => request({ url: `/cart/${id}`, method: 'PUT', data })
export const deleteCart = (ids) => request({ url: '/cart', method: 'DELETE', data: { ids } })
export const selectAllCart = (selected) => request({ url: '/cart/select-all', method: 'PUT', data: { selected } })

// ========== 地址 ==========
export const createAddress = (data) => request({ url: '/address', method: 'POST', data })
export const getAddressList = () => request({ url: '/address' })
export const updateAddress = (id, data) => request({ url: `/address/${id}`, method: 'PUT', data })
export const deleteAddress = (id) => request({ url: `/address/${id}`, method: 'DELETE' })

// ========== 订单 ==========
export const createOrder = (data) => request({ url: '/orders', method: 'POST', data })
export const getOrderList = (params) => request({ url: '/orders', data: params })
export const getOrderDetail = (id) => request({ url: `/orders/${id}` })
export const cancelOrder = (id) => request({ url: `/orders/${id}/cancel`, method: 'PUT' })
export const confirmReceive = (id) => request({ url: `/orders/${id}/receive`, method: 'PUT' })
export const getShipment = (id) => request({ url: `/orders/${id}/shipment` })

// ========== 支付 ==========
export const payOrder = (data) => request({ url: '/pay', method: 'POST', data })
export const refundOrder = (data) => request({ url: '/pay/refund', method: 'POST', data })

// ========== 优惠券 ==========
export const getCouponList = () => request({ url: '/coupons' })
export const receiveCoupon = (id) => request({ url: `/coupons/${id}/receive`, method: 'POST' })
export const getMyCoupons = (params) => request({ url: '/my/coupons', data: params })

// ========== 促销 ==========
export const getPromotions = () => request({ url: '/promotions' })

// ========== 收藏 ==========
export const toggleFavorite = (id) => request({ url: `/favorites/${id}`, method: 'POST' })
export const getFavorites = (params) => request({ url: '/favorites', data: params })

// ========== 浏览记录 ==========
export const getBrowseHistory = (params) => request({ url: '/history', data: params })
export const clearBrowseHistory = () => request({ url: '/history', method: 'DELETE' })

// ========== 评价 ==========
export const createReview = (data) => request({ url: '/reviews', method: 'POST', data })

// ========== 积分 ==========
export const signIn = () => request({ url: '/points/sign', method: 'POST' })
export const getPointsLog = (params) => request({ url: '/points/log', data: params })

// ========== 消息 ==========
export const getMessages = (params) => request({ url: '/messages', data: params })
export const readMessage = (id) => request({ url: `/messages/${id}/read`, method: 'PUT' })
export const readAllMessages = () => request({ url: '/messages/read-all', method: 'PUT' })

// ========== 上传 ==========
export const uploadFile = (filePath) => {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: API_BASE + '/upload',
      filePath,
      name: 'file',
      header: { 'Authorization': `Bearer ${uni.getStorageSync('token')}` },
      success: (res) => {
        const data = JSON.parse(res.data)
        data.code === 0 ? resolve(data.data) : reject(data)
      },
      fail: reject
    })
  })
}
```

---

## 2. 微信小程序登录流程（方案 A）

```javascript
// pages/login/login.vue
import { wxLogin } from '@/utils/api'

export default {
  methods: {
    async handleWxLogin() {
      // 1. 获取微信 code
      const [err, res] = await uni.login({ provider: 'weixin' })
      if (err) return

      // 2. 获取用户信息（需要用户授权按钮）
      const [err2, userInfo] = await uni.getUserProfile({ desc: '用于完善用户资料' })

      // 3. 调后端登录
      const data = await wxLogin({
        code: res.code,
        nickname: userInfo?.userInfo?.nickName || '',
        avatar: userInfo?.userInfo?.avatarUrl || ''
      })

      // 4. 保存 token
      uni.setStorageSync('token', data.token)
      uni.setStorageSync('userInfo', data.user)

      uni.switchTab({ url: '/pages/index/index' })
    }
  }
}
```

---

## 3. 微信支付流程（方案 A）

```javascript
import { createOrder, payOrder } from '@/utils/api'

async function checkout(addressId, cartIds, userCouponId) {
  // 1. 创建订单
  const order = await createOrder({
    address_id: addressId,
    cart_ids: cartIds,
    user_coupon_id: userCouponId  // 可选
  })

  // 2. 获取支付参数
  const openid = uni.getStorageSync('openid')
  const payParams = await payOrder({
    order_id: order.id,
    openid: openid
  })

  // 3. 调起微信支付
  return new Promise((resolve, reject) => {
    uni.requestPayment({
      provider: 'wxpay',
      timeStamp: payParams.timeStamp,
      nonceStr: payParams.nonceStr,
      package: payParams.package,
      signType: payParams.signType,
      paySign: payParams.paySign,
      success: () => {
        uni.showToast({ title: '支付成功' })
        resolve(true)
      },
      fail: (err) => {
        if (err.errMsg.includes('cancel')) {
          uni.showToast({ title: '已取消支付', icon: 'none' })
        }
        reject(err)
      }
    })
  })
}
```

---

## 4. 方案 B：shopxo-uniapp（`api.php`）适配要点

若使用官方 **[shopxo-uniapp](https://github.com/gongfuxiang/shopxo-uniapp)**，前端通过 `get_request_url` 拼接：

`request_url` + `'api'` + `'.php?s='` + `控制器/方法`（见该项目 `App.vue`）。

因此 **`request_url` 必须是站点根地址，且建议以 `/` 结尾**（官方示例：`https://d1.shopxo.vip/`）。设为 `http://host:8080/api` 会导致路径错误（例如拼出 `.../apiapi.php`）。

### 4.1 修改 API 根地址（官方仓库）

在 **`App.vue`** → `globalData.data` 中修改：

```javascript
request_url: 'https://your-goshop-domain.com/',
```

部署上需保证该域名下能访问 GoShop 的 **`/api.php`**（与兼容层路由一致）。未实现或行为不一致的 `s=` 动作需改前端或扩展 **`internal/compat/shopxo`**。

### 4.2 接口路径对照（仅当你重写为方案 A REST 时参考）

下表表示 **业务含义** 上的 REST 对应关系，供**自写/uni 改造为原生 REST** 时查阅；**不是** shopxo-uniapp 默认请求的 URL（默认仍是 `api.php?s=...`）。

| 常见 ShopXO 风格 | GoShop REST（方案 A） | 说明 |
|---|---|---|
| user/login | POST `/api/login` | 账号登录 |
| user/reg | POST `/api/register` | 注册 |
| 微信小程序插件 | POST `/api/wx/login` | 以实际路由为准 |
| goods/index | GET `/api/goods` | 商品列表 |
| goods/detail | GET `/api/goods/:id` | 商品详情 |
| cart/index | GET `/api/cart` | 购物车 |
| buy/add | POST `/api/orders` | 创建订单 |
| order/index | GET `/api/orders` | 订单列表 |
| pay/index | POST `/api/pay` | 发起支付 |

### 4.3 响应格式差异

ShopXO 响应格式：
```json
{"code": 0, "msg": "success", "data": {}}
```

GoShop 响应格式（一致）：
```json
{"code": 0, "msg": "success", "data": {}}
```

响应格式已保持一致，无需额外适配。

### 4.4 价格处理

GoShop 所有价格以「分」存储（int64），前端显示时需要除以 100：

```javascript
// 价格格式化工具
export const formatPrice = (price) => {
  return (price / 100).toFixed(2)
}
```

兼容层返回若仍为「元」等 ShopXO 习惯，以实际 JSON 为准，必要时在方案 B 中单独适配。

### 4.5 图片路径

GoShop 上传的图片路径格式为 `/uploads/.../xxx.jpg`。

- **方案 A**：静态资源域名通常与 API 同源时，可用站点根 + 相对路径，例如 `API_BASE.replace(/\/api\/?$/, '') + path`。
- **方案 B**：与 shopxo-uniapp 的 `request_url` / `static_url` 一致，相对路径拼到站点根（参见官方 `App.vue` 中 `static_url`）。

---

## 5. 开发调试

### 微信开发者工具配置

1. 项目设置 → 不校验合法域名（开发阶段）
2. 详情 → 本地设置 → 勾选「不校验合法域名」

### 真机调试

需要在微信公众平台配置服务器域名：
- request 合法域名：`https://yourdomain.com`
- uploadFile 合法域名：`https://yourdomain.com`

### 常见问题

1. **跨域问题**：GoShop 已内置 CORS 中间件，开发环境无需额外配置
2. **Token 过期**：默认 72 小时，可在 config.yaml 中调整
3. **图片上传失败**：检查 uploads 目录权限，确保可写
4. **支付失败**：确认 config.yaml 中微信支付配置正确，证书文件存在

### APP 拉起小程序收银台（`cashier/paydata`，与 ShopXO 系约定对照）

1. 在后台新增支付方式，JSON 配置中设置 `"payment":"WeixinAppMini"`，可选 `"path":"pages/cashier/cashier"`（与 shopxo-uniapp 常见收银台页路径一致）。名称含「APP小程序」也会被识别为同一模式。
2. 用户在 **APP** 内对订单发起支付且选用该方式时，`order/pay` 在无 `openid` 时会创建 **PayLog**，响应 `data` 为 `weixinapp://...?order_no=<pay_no>`，用于打开小程序。
3. 小程序收银台页调用 `wx.login` 取 `code`，请求 `GET/POST /api.php?s=cashier/paydata`，参数 **`authcode`**（即 code）、**`order_no`**（上一步的 pay_no）。服务端用 `authcode` 换 `openid`，且该 openid 须与订单用户已在库中绑定的微信一致（`user_platforms` 或 `users.open_id`）。
4. 返回结构与 `order/pay` 的线上支付一致（含 `data` 内 `prepay_id` 等），供 `wx.requestPayment` 调起。
