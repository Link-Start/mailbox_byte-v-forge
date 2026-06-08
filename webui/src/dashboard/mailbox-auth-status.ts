import { MailboxAuthStatus } from './dashboard-kit';
import type { Mailbox } from './types';

const authStatusNames = [
  'AUTHORIZED',
  'OAUTH_PENDING',
  'AUTH_FAILED',
  'NEEDS_MANUAL_VERIFICATION',
  'PASSWORD_ONLY',
  'WEBHOOK_ONLY',
  'DISABLED',
  'UNKNOWN'
] as const;

type CanonicalMailboxAuthStatus = typeof authStatusNames[number];

const authStatusEnumByName: Record<CanonicalMailboxAuthStatus, MailboxAuthStatus> = {
  AUTHORIZED: MailboxAuthStatus.MAILBOX_AUTH_STATUS_AUTHORIZED,
  OAUTH_PENDING: MailboxAuthStatus.MAILBOX_AUTH_STATUS_OAUTH_PENDING,
  AUTH_FAILED: MailboxAuthStatus.MAILBOX_AUTH_STATUS_AUTH_FAILED,
  NEEDS_MANUAL_VERIFICATION: MailboxAuthStatus.MAILBOX_AUTH_STATUS_NEEDS_MANUAL_VERIFICATION,
  PASSWORD_ONLY: MailboxAuthStatus.MAILBOX_AUTH_STATUS_PASSWORD_ONLY,
  WEBHOOK_ONLY: MailboxAuthStatus.MAILBOX_AUTH_STATUS_WEBHOOK_ONLY,
  DISABLED: MailboxAuthStatus.MAILBOX_AUTH_STATUS_DISABLED,
  UNKNOWN: MailboxAuthStatus.MAILBOX_AUTH_STATUS_UNKNOWN
};

const authStatusNameByValue = new Map<string, CanonicalMailboxAuthStatus>();
for (const name of authStatusNames) {
  authStatusNameByValue.set(name, name);
  authStatusNameByValue.set(authStatusEnumByName[name], name);
}

export function authStatus(mailbox: Mailbox) {
  return normalizeAuthStatus(String(mailbox.auth_status || '').trim()) || 'OAUTH_PENDING';
}

export function authStatusEnum(mailbox: Mailbox) {
  return authStatusEnumByName[authStatus(mailbox)] || MailboxAuthStatus.MAILBOX_AUTH_STATUS_UNKNOWN;
}

function normalizeAuthStatus(value: string) {
  return authStatusNameByValue.get(value) || '';
}
