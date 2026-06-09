import PostalMime from "postal-mime";
import { convert } from "html-to-text";
import { clean, cleanHeader, normalizeEmail, receivedAt, stableEventId, uniqueEmails } from "./text.js";

export async function buildEmailEvent(message) {
  const parsed = await PostalMime.parse(message.raw);
  const textBody = emailTextBody(parsed);
  const recipients = emailRecipients(message, parsed);
  const messageId = cleanHeader(parsed.messageId || message.headers.get("message-id") || "");
  const receivedAtUnix = receivedAt(parsed.date || message.headers.get("date"));
  return {
    version: 1,
    provider: "cloudflare",
    eventId: await stableEventId("cloudflare", messageId, message.from, recipients.join(","), receivedAtUnix),
    messageId,
    fromAddress: normalizeEmail(parsed.from?.address || message.from || ""),
    recipients,
    subject: cleanHeader(parsed.subject || message.headers.get("subject") || ""),
    textBody,
    htmlBody: String(parsed.html || ""),
    receivedAtUnix,
    rawSize: Number(message.rawSize || 0),
  };
}

function emailTextBody(parsed) {
  return clean(
    parsed.text ||
      convert(parsed.html || "", {
        wordwrap: false,
        selectors: [
          { selector: "img", format: "skip" },
          { selector: "style", format: "skip" },
          { selector: "script", format: "skip" },
          { selector: "head", format: "skip" },
          { selector: "a", options: { ignoreHref: true } },
        ],
      }),
  );
}

function emailRecipients(message, parsed) {
  return uniqueEmails([
    message.to,
    ...addressList(parsed.to),
    ...addressList(parsed.cc),
    ...addressList(parsed.bcc),
  ]);
}

function addressList(values) {
  if (!Array.isArray(values)) return [];
  return values.map((item) => item?.address || "");
}
