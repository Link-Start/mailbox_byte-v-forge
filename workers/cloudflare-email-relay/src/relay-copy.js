import { normalizeEmail, uniqueEmails } from "./text.js";

const copyTargetsEnv = "MAILBOX_RELAY_COPY_TO_EMAILS";

export async function forwardEmailCopies(message, env) {
  const targets = emailCopyTargets(env);
  for (const target of targets) {
    await forwardEmailCopy(message, target);
  }
}

function emailCopyTargets(env) {
  return uniqueEmails(String(env[copyTargetsEnv] || "").split(/[\s,;]+/));
}

async function forwardEmailCopy(message, target) {
  const email = normalizeEmail(target);
  if (!email) return;
  try {
    await message.forward(email);
    console.log(`forwarded email copy target=${redactEmail(email)}`);
  } catch (error) {
    console.error(`forward email copy failed target=${redactEmail(email)} error=${errorMessage(error)}`);
  }
}

function redactEmail(email) {
  const [local, domain] = normalizeEmail(email).split("@");
  if (!local || !domain) return "<email>";
  return `${local.slice(0, 2)}***@${domain}`;
}

function errorMessage(error) {
  return error instanceof Error ? error.message : String(error);
}
