# 反向代理与支付 / DIY — 手工验收清单

> 文中「ShopXO uni-app」指常见 shopxo-uniapp 请求路径，仅作验收场景说明。

在**已配置 HTTPS 终止**（Nginx/Caddy 等）的真实域名下，按顺序粗测即可。Nginx 需转发：

- `X-Forwarded-Proto`（通常为 `$scheme`）
- 建议同时转发 `X-Forwarded-Host`、`X-Forwarded-For`

## ShopXO uni-app `order/pay`

1. **单笔**：待付款订单，`ids`（或 `id`）仅一个，`payment_id` 为后台启用的支付方式 ID；线下/钱包应直接成功；微信支付宝在 `payment.sandbox: true` 时应出现沙盒回调 URL。
2. **多笔合并**：勾选两笔及以上待付款订单，同一 `payment_id` 发起支付；线下应对多笔一并置为已支付；沙盒模式下应对 `PayLog.pay_no` 触发 `/api/pay/sandbox/callback` 后全部已支付。

## 支付回调

- 微信/支付宝生产回调应能根据 **`order_no`（单笔）** 或 **`pay_no`（合并）** 将订单置为已付（见 `HandlePayNotify`）。

## DIY / form 管理端

1. 打开依赖 `diyapi/init` 或 `forminputapi/init` 的页面，检查返回的 `attachment_host`、`public_host` 是否为 **`https://你的域名`**（而非 `http://`）。
2. **附件远程抓取**：管理员登录后请求 `POST /api/attachmentapi/catch`（body 含 `source` 为可公网访问的图片 URL、`category_id` 可选），确认 `data` 中返回 `Attachment` 且文件落在 `uploads/` 下。

## 数据库

- 新环境首次启动后确认 `orders.payment_id` 列存在；若关闭 AutoMigrate，执行 `docs/deployment.md` 中的 `ALTER TABLE` 示例。
