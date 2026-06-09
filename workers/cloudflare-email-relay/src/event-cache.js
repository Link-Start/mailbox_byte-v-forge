import { normalizeEmail, positiveInt } from "./text.js";

const CACHE_PREFIX = "pending:v1:";
const DEFAULT_CACHE_TTL_SECONDS = 3 * 24 * 60 * 60;
const MAX_CACHE_TTL_SECONDS = 7 * 24 * 60 * 60;
const DEFAULT_PULL_LIMIT = 50;
const MAX_PULL_LIMIT = 200;
const MAX_SCAN_LIMIT = 1000;

export async function cacheEmailEvent(env, event) {
  const cache = emailEventCache(env);
  if (!cache) return false;
  try {
    await cache.put(cacheKey(event.eventId), JSON.stringify(event), {
      expirationTtl: cacheTTLSeconds(env),
    });
    return true;
  } catch (error) {
    console.error(`email event cache write failed event_id=${event.eventId} error=${errorMessage(error)}`);
    return false;
  }
}

export async function deleteCachedEmailEvents(env, eventIds) {
  const cache = emailEventCache(env);
  if (!cache) return 0;
  const ids = uniqueEventIds(eventIds);
  await Promise.all(ids.map((eventId) => cache.delete(cacheKey(eventId))));
  return ids.length;
}

export async function listCachedEmailEvents(env, options = {}) {
  const cache = emailEventCache(env);
  if (!cache) throw new Error("EMAIL_EVENT_CACHE binding is required");
  const limit = positiveInt(options.limit, DEFAULT_PULL_LIMIT, MAX_PULL_LIMIT);
  const scanLimit = Math.min(Math.max(limit * 4, limit), MAX_SCAN_LIMIT);
  const listed = await cache.list({ prefix: CACHE_PREFIX, limit: scanLimit });
  const rows = await Promise.all(listed.keys.map((key) => readCachedEmailEvent(cache, key.name)));
  const recipient = normalizeEmail(options.recipient);
  return rows
    .filter((event) => event && recipientMatches(event, recipient))
    .sort((left, right) => Number(right.receivedAtUnix || 0) - Number(left.receivedAtUnix || 0))
    .slice(0, limit);
}

export async function ackCachedEmailEvents(env, eventIds) {
  return deleteCachedEmailEvents(env, eventIds);
}

export function eventIdsFromAckBody(body) {
  return uniqueEventIds([...(body?.event_ids || []), ...(body?.eventIds || [])]).slice(0, MAX_PULL_LIMIT);
}

export function pullLimit(value) {
  return positiveInt(value, DEFAULT_PULL_LIMIT, MAX_PULL_LIMIT);
}

function emailEventCache(env) {
  const cache = env.EMAIL_EVENT_CACHE;
  return cache && typeof cache.put === "function" && typeof cache.get === "function" ? cache : null;
}

function cacheTTLSeconds(env) {
  return positiveInt(env.EMAIL_EVENT_CACHE_TTL_SECONDS, DEFAULT_CACHE_TTL_SECONDS, MAX_CACHE_TTL_SECONDS);
}

function cacheKey(eventId) {
  return `${CACHE_PREFIX}${eventId}`;
}

async function readCachedEmailEvent(cache, key) {
  try {
    const value = await cache.get(key, "json");
    return value && value.eventId ? value : null;
  } catch (error) {
    console.error(`email event cache read failed key=${key} error=${errorMessage(error)}`);
    return null;
  }
}

function recipientMatches(event, recipient) {
  if (!recipient) return true;
  return (event.recipients || []).some((value) => normalizeEmail(value) === recipient);
}

function uniqueEventIds(values) {
  const seen = new Set();
  const out = [];
  for (const value of values || []) {
    const eventId = String(value || "").trim();
    if (!eventId || seen.has(eventId)) continue;
    seen.add(eventId);
    out.push(eventId);
  }
  return out;
}

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error);
}
