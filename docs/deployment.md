# GoShop 部署指南

## 方式一：Docker Compose（推荐）

```bash
git clone https://github.com/zhangpanda/goshop.git
cd goshop
cp config.yaml.example config.yaml
```

编辑 `config.yaml`，将 `db.host` 改为 `mysql`，`db.password` 改为 `goshop123`：

```yaml
db:
  host: mysql
  password: "goshop123"
# Redis 可选，配置后启用，留空则使用内存缓存
redis:
  host: redis    # 不需要 Redis 可留空
```

启动：

```bash
docker compose up -d
```

服务地址：
- 后端 API：http://localhost:8080
- 默认管理员：admin / admin123

## 方式二：手动部署

### 环境要求

- Go 1.23+
- MySQL 8.0+
- Redis 7+（可选，不配置则使用内存缓存）
- Node.js 20+（构建前端）

### 1. 编译后端

```bash
cd goshop
go build -o bin/goshop ./cmd/server/
```

### 2. 构建 PC 前台

```bash
cd web
npm ci
npm run build
```

### 3. 构建管理后台

```bash
cd admin
npm ci
npm run build
```

### 4. 配置

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml 填写数据库连接信息
# Redis 可选：留空 redis.host 则自动使用内存缓存
# 生产环境务必修改 jwt.secret
```

### 5. 启动

```bash
# 后端
./bin/goshop

# PC 前台（生产模式）
cd web && npm start

# 管理后台（生产模式）
cd admin && npm start
```

## 方式三：Systemd 服务

创建 `/etc/systemd/system/goshop.service`：

```ini
[Unit]
Description=GoShop Backend
After=network.target mysql.service

[Service]
Type=simple
User=www
WorkingDirectory=/opt/goshop
ExecStart=/opt/goshop/bin/goshop
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable goshop
sudo systemctl start goshop
```

## Nginx 反向代理

```nginx
# 后端 API + 静态文件
server {
    listen 80;
    server_name api.yourdomain.com;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# PC 前台
server {
    listen 80;
    server_name www.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# 管理后台
server {
    listen 80;
    server_name admin.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 数据库：`orders.payment_id`（关闭 AutoMigrate 时）

默认启动会执行 GORM `AutoMigrate`，会自动为 `orders` 表增加 `payment_id`（`uint`，默认 0，索引）。

若生产环境**禁止自动迁移**，请手动执行与模型一致的 DDL（MySQL 示例）：

```sql
ALTER TABLE `orders`
  ADD COLUMN `payment_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '支付方式ID(用户选用)' AFTER `remark`,
  ADD INDEX `idx_orders_payment_id` (`payment_id`);
```

若列已存在，可省略本段。字段语义见 `internal/model/order.go` 中 `Order.PaymentID`。

## 生产环境检查清单

- [ ] `config.yaml` 中 `server.mode` 改为 `release`
- [ ] `jwt.secret` 使用强随机字符串（32位以上）
- [ ] MySQL 密码使用强密码
- [ ] Redis 设置密码（如启用）
- [ ] 配置 HTTPS（Let's Encrypt）
- [ ] 配置防火墙，只开放 80/443 端口
- [ ] 定期备份 MySQL 数据库
- [ ] 配置日志轮转

## 微信支付配置

```yaml
wechat:
  app_id: "your_appid"              # 小程序 AppID
  app_secret: "your_appsecret"      # 小程序 AppSecret
  mch_id: "your_mchid"              # 商户号
  mch_api_key: "your_v3_api_key"    # APIv3 密钥
  serial_no: "your_cert_serial_no"  # 商户证书序列号
  private_key: "cert/apiclient_key.pem"  # 商户私钥文件路径
  notify_url: "https://api.yourdomain.com/api/pay/notify"  # 支付回调地址
```

将商户证书文件放在 `cert/` 目录下。

## 支付宝支付配置

```yaml
alipay:
  app_id: "your_alipay_appid"       # 支付宝应用 AppID
  private_key: "your_private_key"    # 应用私钥（RSA2, PKCS8 格式）
  public_key: "your_alipay_pubkey"   # 支付宝公钥（用于验签回调）
  notify_url: "https://api.yourdomain.com/api/pay/alipay-notify"
```

> 私钥和公钥可在[支付宝开放平台](https://open.alipay.com)的应用设置中获取。
