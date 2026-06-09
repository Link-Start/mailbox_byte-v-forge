import { clip } from "./text.js";

const WEBHOOK_TIMEOUT_MS = 8000;
const WEBHOOK_ATTEMPTS = 3;

export async function forwardEmailEvent(env, event) {
  const url = String(env.MAILBOX_WEBHOOK_URL || "").trim();
  const token = String(env.MAILBOX_WEBHOOK_TOKEN || "").trim();
  if (!url || !token) {
    console.error("MAILBOX_WEBHOOK_URL and MAILBOX_WEBHOOK_TOKEN are required");
    return false;
  }
  const body = JSON.stringify(event);
  for (let attempt = 1; attempt <= WEBHOOK_ATTEMPTS; attempt += 1) {
    if (await postEmailEvent(url, token, body, attempt)) return true;
    if (attempt < WEBHOOK_ATTEMPTS) await sleep(attempt * 250);
  }
  return false;
}

async function postEmailEvent(url, token, body, attempt) {
  try {
    const response = await fetchWithTimeout(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Webhook-Token": token,
      },
      body,
    });
    if (response.ok) return true;
    console.error(`mailbox webhook failed attempt=${attempt} status=${response.status} body=${clip(await response.text(), 300)}`);
    return false;
  } catch (error) {
    console.error(`mailbox webhook failed attempt=${attempt} error=${errorMessage(error)}`);
    return false;
  }
}

async function fetchWithTimeout(url, init) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort("timeout"), WEBHOOK_TIMEOUT_MS);
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
