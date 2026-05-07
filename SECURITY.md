# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| main    | ✅        |

## Reporting a Vulnerability

**请勿通过公开 Issue 报告安全漏洞。**

如果你发现安全问题，请通过以下方式私下报告：

- Email: security@goshop.dev
- GitHub Security Advisories: https://github.com/zhangpanda/goshop/security/advisories/new

我们会在 72 小时内确认收到，并在 7 个工作日内给出初步评估。

## 已知安全边界

- **SQL Console**：默认关闭（`server.sql_console: false`），仅超级管理员可用，仅允许 SELECT/SHOW/DESC/EXPLAIN
- **限流**：进程内滑动窗口，多实例部署需在网关层补充
- **CORS**：未配置 `cors_origins` 时仅允许 localhost，生产需同域反代或显式配置白名单
- **默认管理员**：首次启动创建 `admin/admin123`，**部署后必须立即修改密码**
