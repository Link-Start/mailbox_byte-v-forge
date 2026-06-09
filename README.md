# mailbox

`mailbox` 是邮箱能力仓，负责邮箱账号、provider adapter、注册/OAuth 编排、收件、邮件信号解析和独立 dashboard。

## 核心能力

- 提供 Mailbox gRPC API、Dashboard HTTP API 和自带静态 Web UI。
- 内置 Outlook 注册/OAuth、Microsoft Graph 收件和 Cloudflare Email Routing webhook 链路。
- 解析入站邮件中的验证码等可复用信号，并以 secret ref/artifact ref 对外暴露。
- 支持 mailbox 自有事件、operation 投影、收件缓存和跨副本抓取协调。
- 通过 provider capability 描述 Outlook、Cloudflare 等能力差异，避免业务侧硬编码 provider 细节。

## 使用方式

业务服务通过 mailbox API、事件或查询读取邮箱状态与邮件信号；真实凭据、provider raw shape 和内部 operation 状态只留在 mailbox 内部。Outlook 浏览器动作可接入 `browser-automation`，Cloudflare 入站邮件通过 webhook 或 relay pull 进入本服务。

## 入口

- 服务入口：`services/mailbox-api/`
- Cloudflare relay worker：`workers/cloudflare-email-relay/`
- 契约真源：`proto/`
- Dashboard：`webui/`

## 常用检查

```sh
sh scripts/generate-proto.sh
git diff --check
```
