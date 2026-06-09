import { ackCachedEmailEvents, eventIdsFromAckBody, listCachedEmailEvents, pullLimit } from "./event-cache.js";
import { normalizeEmail } from "./text.js";

const relayTokenHeader = "X-Relay-Token";

export async function handleRelayAPI(request, env) {
  const url = new URL(request.url);
  if (url.pathname === "/healthz") return new Response(null, { status: 204 });
  if (url.pathname === "/pending") return handlePending(request, env, url);
  if (url.pathname === "/ack") return handleAck(request, env);
  return null;
}

async function handlePending(request, env, url) {
  if (request.method !== "GET") return methodNotAllowed("GET");
  if (!authorized(request, env)) return json({ error_message: "unauthorized" }, 401);
  try {
    const events = await listCachedEmailEvents(env, {
      limit: pullLimit(url.searchParams.get("limit")),
      recipient: normalizeEmail(url.searchParams.get("recipient")),
    });
    return json({ events });
  } catch (error) {
    return json({ error_message: errorMessage(error) }, 503);
  }
}

async function handleAck(request, env) {
  if (request.method !== "POST") return methodNotAllowed("POST");
  if (!authorized(request, env)) return json({ error_message: "unauthorized" }, 401);
  const body = await request.json().catch(() => ({}));
  const acked = await ackCachedEmailEvents(env, eventIdsFromAckBody(body));
  return json({ acked_count: acked });
}

function authorized(request, env) {
  const expected = String(env.MAILBOX_RELAY_PULL_TOKEN || "").trim();
  if (!expected) return false;
  const headerToken = String(request.headers.get(relayTokenHeader) || "").trim();
  const bearerToken = String(request.headers.get("Authorization") || "").replace(/^Bearer\s+/i, "").trim();
  return headerToken === expected || bearerToken === expected;
}

function methodNotAllowed(method) {
  return new Response("method not allowed", { status: 405, headers: { Allow: method } });
}

function json(value, status = 200) {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error);
}
