# mailbox

Mailbox 领域仓，承载邮箱账号、Outlook provider、邮箱注册/OAuth MQ 编排和收件能力。

## 目录

- `services/mailbox-api`：Mailbox 领域 gRPC API 和唯一服务进程，内置 Outlook/Cloudflare provider adapter、邮箱注册/OAuth MQ worker、收件、webhook 和邮件信号解析能力。
- `Dockerfile`：独立部署入口，构建 mailbox API 与内置 dashboard 静态资源，只启动 mailbox 一个服务进程。
- `workers/cloudflare-email-relay`：Cloudflare Email Routing Worker，将 CF 入站邮件转发到 mailbox webhook。
- `proto/email.proto`：邮件读取服务契约。
- `proto/mailbox_register.proto`：邮箱注册与 OAuth 编排模型。
- `proto/mailbox_service.proto`：Mailbox 领域 API 契约。
- `proto/byte/v/forge/contracts/...`：mailbox 独立部署所需的 mailbox/common/observability/browserautomation 契约副本，Go 生成物进入 `services/mailbox-api/internal/contracts`。

Mailbox 后端不再依赖 `common-lib` Go module，webui 也不再依赖 `@byte-v-forge/common-ui`；服务 API、事件、provider capability、domain、browser-automation 客户端和 dashboard 类型均由本仓 proto 生成。`mailbox` 仓内部模型可以保存 password、refresh/access token 等拥有方细节；对外邮箱查询只暴露 credential state，不回显凭据值；对外收件结果只暴露 secret ref 和 artifact ref。

## 生成

```sh
sh scripts/generate-proto.sh
```

生成物用于远程构建和部署验证，位于仓库忽略路径。

## 配置

`services/mailbox-api` 直接内置 Outlook 和 Cloudflare provider adapter。配置 `MAILBOX_PG_DSN` 时使用 PostgreSQL 维护邮箱、邮件、operation 和平台事件 outbox 投影；未配置时使用本进程内存仓储，适合 standalone 试用，进程重启后邮箱、邮件和 operation 状态不会保留。

Dashboard 由 mailbox 服务自身托管，不再发布 Module Federation remote。静态资源默认从 `MAILBOX_DASHBOARD_STATIC_DIR=/app/dashboard/mailbox` 读取，服务在 `/dashboard/mailbox/` 和根路径提供独立 SPA，在 `/api/mailbox/*` 提供 dashboard BFF API。

Outlook 注册和 OAuth 通过 mailbox 内置 worker 编排：`RegisterMailbox` / `RunMailboxOAuth` 只创建 operation；配置 `MAILBOX_NATS_URL` 时发布 `mailbox.registration.operation_requested` / `mailbox.oauth.operation_requested` 到 mailbox JetStream，由 mailbox worker 消费后调用可选 `browser-automation` 执行并更新 operation 投影；未配置 MQ 时由本进程 local worker 执行。claim owner、lease、attempt count、OAuth limit/only_missing 只保存在 mailbox 内部表中，不进入 dashboard/API 对外 `MailboxOperation` 模型。

`MAILBOX_BROWSER_AUTOMATION_ADDR` 是可选配置；未配置时 mailbox 仍可启动，Outlook 注册/OAuth 浏览器动作会返回不可用错误，收件、Cloudflare webhook 和 dashboard 继续工作。配置该地址后，Outlook 注册/OAuth 浏览器 profile 通过 `OUTLOOK_REGISTER_AUTOMATION_PROXY_REF`、`OUTLOOK_REGISTER_AUTOMATION_LOCALE` 和 `OUTLOOK_REGISTER_AUTOMATION_TIMEZONE` 控制。

Outlook 邮件读取使用 Microsoft Graph Go SDK 读取当前 OAuth 用户的 messages，并用 `Prefer: outlook.body-content-type="text"` 请求文本正文；不再保留手写 Graph REST adapter 或额外 URL 覆盖。

Cloudflare 邮件是主动推送链路：Email Routing Worker 收到邮件后把标准化事件 POST 到 `/webhooks/email/cloudflare`，mailbox 服务使用 `MAILBOX_WEBHOOK_HTTP_ADDR` 开启 HTTP webhook，并只通过 `X-Webhook-Token` 读取 `MAILBOX_WEBHOOK_TOKEN` 校验转发请求。Outlook Graph webhook 使用 `/webhooks/email/microsoft-graph`，验证 URL 必须带同一个 token，POST 通知的 `clientState` 也必须等于同一个 token。Cloudflare 域名池来自 Cloudflare API：`MAILBOX_CLOUDFLARE_API_TOKEN` 读取 `MAILBOX_CLOUDFLARE_EMAIL_CONFIG_FILE` 中声明的 zones，并从 Email Routing catch-all 规则与 Cloudflare MX DNS 记录推导可用邮箱域名；token 限制到目标 zone，并授予 `Zone Read`、`DNS Read` 和 `Email Routing Rules Read` 即可。Cloudflare 地址不需要手动导入，邮件到达后按 recipient 自动形成虚拟邮箱并按 domain 分组展示。需要公网入口时在 deploy 的 `ingress.webhook` 暴露 mailbox webhook，或使用受管 HTTPS 隧道把该入口映射到公网域名；业务代码不管理公网隧道。

邮件保留策略按 provider 独立执行 FIFO：Outlook 使用 `MAILBOX_OUTLOOK_MAX_MESSAGES_PER_MAILBOX` 限定每个邮箱的最大邮件数，Cloudflare 使用 `MAILBOX_CLOUDFLARE_MAX_MESSAGES_PER_DOMAIN` 限定每个 domain 的最大邮件数。超过上限时删除最早邮件及对应 seen 记录，默认分别为 `100` 和 `500`。

邮件内容会先落库，再通过 mailbox 通用解析器生成 `EmailSignal`。通用解析器只识别验证码等可复用邮件信号；验证码原文写入 mailbox 自有 Redis TTL secret store，对外 `EmailSignal.secret_ref` 只返回可解析引用与过期时间，dashboard 不展示或复制验证码原文。业务状态判断由业务服务通过 webhook 或查询读取邮件后自行完成。

Redis 是可选运行增强而非启动前置条件。配置 `MAILBOX_RECENT_EMAIL_REDIS_URL` 后，新入库邮件按邮箱写入近期热缓存，并用同一 TTL 维护邮箱验证码 secret；未配置时跳过近期缓存和 TTL secret store，`WaitForMailboxEmail` 直接回查 mailbox 仓储投影。配置 `MAILBOX_COORDINATION_REDIS_URL` 后，`MAILBOX_INBOX_LOCK_KEY_PREFIX` 用于跨副本抓取锁和 Outlook webhook refresh 锁；未配置时按单副本 standalone 模式无分布式锁执行。UI 实时刷新通过 HotStream/NATS Core 或进程内 HotStream 的非持久化通知触发前端重新查询。Redis 仅作为热点读取、短期 secret 与协调层，不作为邮件领域状态真源。

同时配置 `MAILBOX_PG_DSN` 和 `MAILBOX_NATS_URL` 时，邮件入库会在同一 DB transaction 写入 `mailbox_event_outbox`，再由 outbox worker 发布公共 `mailbox.email.received` / `mailbox.email.signal.received` 事件，事件流名称由 `MAILBOX_EVENT_STREAM_NAME` 指定，默认 `MAILBOX_EVENTS`，subject 使用 `mailbox.>`；启动时会通过 NATS SDK 确保 JetStream stream 存在，避免邮件已落库但 NATS 临时失败导致下游投影丢失；`FetchMailboxInboxes`、注册/OAuth 和入站 poll 通过 mailbox command event worker 异步执行。未配置 `MAILBOX_NATS_URL` 或未配置 `MAILBOX_PG_DSN` 时不使用 DB outbox 派发 command，注册/OAuth/fetch operation 由 mailbox 本进程 local worker goroutine 执行，`WaitForMailboxEmail` 直接触发本地 Outlook poll。业务服务需要消费邮箱事件时应订阅 mailbox events 并在自身服务内维护幂等投影，不再通过 mailbox outbound HTTP webhook 旁路投递。

## 检查

```sh
cd ../deploy
./scripts/deploy-remote.sh mailbox
```

业务构建、镜像构建和部署验证统一在远程宿主机执行，本机只做源码编辑和调度。
