# GoShop 部署指南

## 方式一：Docker Compose（推荐）

```bash
git clone https://github.com/zhangpanda/goshop.git
cd goshop
cp config.yaml.example config.yaml
```

编辑 `config.yaml`，将 `db.host` 改为 `mysql`，`redis.host` 改为 `redis`，`db.password` 改为 `goshop123`：

```yaml
db:
  host: mysql
  password: "goshop123"
redis:
  host: redis
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
- Redis 7+
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
# 编辑 config.yaml 填写实际的数据库和 Redis 连接信息
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
After=network.target mysql.service redis.service

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

## 生产环境检查清单

- [ ] `config.yaml` 中 `server.mode` 改为 `release`
- [ ] `jwt.secret` 使用强随机字符串（32位以上）
- [ ] MySQL 密码使用强密码
- [ ] Redis 设置密码
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
