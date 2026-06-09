import { buildEmailEvent } from "./email-event.js";
import { cacheEmailEvent, deleteCachedEmailEvents } from "./event-cache.js";
import { forwardEmailCopies } from "./relay-copy.js";
import { forwardEmailEvent } from "./relay-forward.js";
import { handleRelayAPI } from "./relay-api.js";

export default {
  async fetch(request, env) {
    const response = await handleRelayAPI(request, env);
    return response || new Response("cloudflare-email-relay worker is running", { status: 200 });
  },

  async email(message, env) {
    const event = await buildEmailEvent(message);
    const cached = await cacheEmailEvent(env, event);
    const forwarded = await forwardEmailEvent(env, event);
    if (forwarded) {
      await deleteCachedEmailEvents(env, [event.eventId]);
    } else if (!cached && env.WEBHOOK_FAIL_OPEN !== "true") {
      message.setReject("mailbox webhook delivery failed");
      return;
    }
    await forwardEmailCopies(message, env);
  },
};
