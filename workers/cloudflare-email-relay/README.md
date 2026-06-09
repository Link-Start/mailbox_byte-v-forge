# cloudflare-email-relay

Cloudflare Email Routing Worker that parses inbound email, caches it briefly, and forwards normalized events to mailbox.

## Runtime config

- `MAILBOX_WEBHOOK_URL`: public mailbox webhook endpoint, ending with `/webhooks/email/cloudflare`.
- `MAILBOX_WEBHOOK_TOKEN`: secret sent as `X-Webhook-Token`.
- `WEBHOOK_FAIL_OPEN`: when `true`, accepts the email even if webhook forwarding and cache write both fail.
- `EMAIL_EVENT_CACHE`: KV binding used as pending-email cache for pull compensation.
- `EMAIL_EVENT_CACHE_TTL_SECONDS`: pending cache TTL, default 259200 seconds.
- `MAILBOX_RELAY_PULL_TOKEN`: optional secret required by `GET /pending` and `POST /ack`; defaults to `MAILBOX_WEBHOOK_TOKEN`.
- `TELEGRAM_BOT_TOKEN`: optional Telegram bot token used to send an email notification copy.
- `TELEGRAM_CHAT_ID`: optional Telegram chat ID that receives email notifications.

## Cache and pull API

Create a KV namespace, bind it as `EMAIL_EVENT_CACHE`, and set the pull token:

```sh
wrangler kv namespace create EMAIL_EVENT_CACHE
wrangler secret put MAILBOX_WEBHOOK_TOKEN
wrangler secret put TELEGRAM_BOT_TOKEN
wrangler secret put TELEGRAM_CHAT_ID
```

The Worker stores each parsed email before forwarding. If mailbox webhook delivery succeeds, the cached event is deleted. If the tunnel/webhook is unavailable, mailbox can later pull pending events:

- `GET /pending?recipient=<email>&limit=50` with `X-Relay-Token` returns `{ "events": [...] }`.
- `POST /ack` with `X-Relay-Token` and `{ "event_ids": [...] }` deletes delivered events.

## Run

```sh
npm install
npm run deploy
```

The Worker is the active inbound path for Cloudflare mailboxes. Cloudflare invokes the `email()` handler, this worker parses the MIME message with `postal-mime`, stores a temporary pending copy in KV, then POSTs the normalized event to mailbox. The mailbox service persists the message and exposes generic email signals for downstream services.
