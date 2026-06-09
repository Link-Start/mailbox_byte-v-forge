import { clip, clean, normalizeEmail } from "./text.js";

const TELEGRAM_TIMEOUT_MS = 6000;
const TELEGRAM_ATTEMPTS = 2;
const TELEGRAM_MESSAGE_LIMIT = 3900;
const TELEGRAM_BODY_LIMIT = 3000;

export async function notifyTelegramEmailEvent(env, event) {
  const config = telegramConfig(env);
  if (!config.enabled) return;
  const body = JSON.stringify({
    chat_id: config.chatId,
    text: telegramMessageText(event),
    disable_web_page_preview: true,
  });
  for (let attempt = 1; attempt <= TELEGRAM_ATTEMPTS; attempt += 1) {
    if (await sendTelegramMessage(config.token, body, attempt)) return;
    if (attempt < TELEGRAM_ATTEMPTS) await sleep(attempt * 300);
  }
}

function telegramConfig(env) {
  const token = String(env.TELEGRAM_BOT_TOKEN || "").trim();
  const chatId = String(env.TELEGRAM_CHAT_ID || "").trim();
  return { enabled: Boolean(token && chatId), token, chatId };
}

function telegramMessageText(event) {
  return clip(
    [
      "📬 Mailbox 收到新邮件",
      `主题：${clean(event.subject) || "-"}`,
      `发件人：${normalizeEmail(event.fromAddress) || "-"}`,
      `收件人：${recipientText(event)}`,
      `时间：${receivedText(event)}`,
      "",
      clip(clean(event.textBody || event.htmlBody), TELEGRAM_BODY_LIMIT) || "无文本正文",
    ].join("\n"),
    TELEGRAM_MESSAGE_LIMIT,
  );
}

function recipientText(event) {
  const recipients = Array.isArray(event.recipients) ? event.recipients.map(normalizeEmail).filter(Boolean) : [];
  return recipients.length > 0 ? recipients.join(", ") : "-";
}

function receivedText(event) {
  const receivedAt = Number(event.receivedAtUnix || 0);
  if (receivedAt <= 0) return "-";
  return new Date(receivedAt * 1000).toISOString();
}

async function sendTelegramMessage(token, body, attempt) {
  try {
    const response = await fetchWithTimeout(`https://api.telegram.org/bot${token}/sendMessage`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body,
    });
    if (response.ok) return true;
    console.error(`telegram notify failed attempt=${attempt} status=${response.status}`);
    return false;
  } catch (error) {
    console.error(`telegram notify failed attempt=${attempt} error=${errorMessage(error)}`);
    return false;
  }
}

async function fetchWithTimeout(url, init) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort("timeout"), TELEGRAM_TIMEOUT_MS);
  try {
    return await fetch(url, { ...init, signal: controller.signal });
  } finally {
    clearTimeout(timeout);
  }
}

function sleep(milliseconds) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error);
}
