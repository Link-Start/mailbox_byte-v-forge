export function clean(value) {
  return String(value || "")
    .replace(/\r\n?/g, "\n")
    .replace(/[\u200B-\u200D\uFEFF]/g, "")
    .replace(/\u00A0/g, " ")
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .join("\n")
    .replace(/\n{3,}/g, "\n\n")
    .trim();
}

export function cleanHeader(value) {
  return clean(value).replace(/\n/g, " ");
}

export function uniqueEmails(values) {
  const out = [];
  const seen = new Set();
  for (const value of values) {
    const email = normalizeEmail(value);
    if (!email || seen.has(email)) continue;
    seen.add(email);
    out.push(email);
  }
  return out;
}

export function normalizeEmail(value) {
  return String(value || "").trim().toLowerCase();
}

export function receivedAt(value) {
  const parsed = Date.parse(value || "");
  if (Number.isFinite(parsed)) return Math.floor(parsed / 1000);
  return Math.floor(Date.now() / 1000);
}

export async function stableEventId(...parts) {
  const bytes = new TextEncoder().encode(parts.join("\0"));
  const digest = await crypto.subtle.digest("SHA-256", bytes);
  return [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

export function positiveInt(value, fallback, max) {
  const parsed = Number.parseInt(String(value || ""), 10);
  let out = Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
  if (max > 0 && out > max) out = max;
  return out;
}

export function clip(value, limit) {
  const text = String(value || "");
  return text.length > limit ? `${text.slice(0, limit)}...` : text;
}
