import { MailboxAuthStatus, MailboxCredentialKind, MailboxProviderAction } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { mailboxProviderMatches, mailboxProviderValue, normalizeMailboxProviderKey, type MailboxProviderTab } from './mailbox-provider-config';
import type { Mailbox, MailboxProviderActionCapability, MailboxProviderCapability } from './types';

export { mailboxProviderMatches, mailboxProviderValue, normalizeMailboxProviderKey, type MailboxProviderTab };
export type MailboxBatchItem = {
  email: string;
  password: string;
};

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

export function domainForEmail(email: string) {
  const [, domain = ''] = normalizeUiEmail(email).split('@');
  return domain;
}

export function tokenText(mailbox: Mailbox) {
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN) && authStatus(mailbox) === 'AUTHORIZED') return 'Refresh 可用';
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN)) return 'Refresh 待验证';
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN)) return '仅 Access';
  return '缺 Token';
}

export function authStatus(mailbox: Mailbox) {
  const value = normalizeAuthStatus(String(mailbox.auth_status || '').trim());
  if (value) return value;
  return 'OAUTH_PENDING';
}

export function authStatusEnum(mailbox: Mailbox) {
  return authStatusEnumByName[authStatus(mailbox)] || MailboxAuthStatus.MAILBOX_AUTH_STATUS_UNKNOWN;
}

export function parseMailboxBatch(value: string, credentialKinds: MailboxCredentialKind[]) {
  const items: MailboxBatchItem[] = [];
  const errors: string[] = [];
  const allowPlainEmailBatch = !credentialKinds.includes(MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD);

  value.split(/\r?\n/).forEach((raw, index) => {
    const line = raw.trim();
    if (!line) return;
    const delimiterIndex = line.indexOf('----');
    if (allowPlainEmailBatch && delimiterIndex < 0) {
      items.push({ email: line, password: '' });
      return;
    }
    if (delimiterIndex < 0) {
      errors.push(`第 ${index + 1} 行缺少 ----`);
      return;
    }
    const email = line.slice(0, delimiterIndex).trim();
    const password = line.slice(delimiterIndex + 4).trim();
    if (!email) {
      errors.push(`第 ${index + 1} 行缺少账号`);
      return;
    }
    items.push({ email, password });
  });

  return { items, errors };
}

export function capabilityForProvider(capabilities: MailboxProviderCapability[], provider: string) {
  const target = normalizeMailboxProviderKey(provider);
  if (!target) return undefined;
  return capabilities.find((capability) => mailboxProviderMatches(capability.key, target));
}

export function providerAction(capability: MailboxProviderCapability | undefined, action: MailboxProviderAction) {
  return (capability?.actions || []).find((item) => item.action === action);
}

export function canRunMailboxAction(mailbox: Mailbox, action: MailboxProviderActionCapability | undefined) {
  if (!action) return false;
  if (!requiredCredentialsPresent(mailbox, action.required_credentials || [])) return false;
  const statuses = action.required_auth_statuses || [];
  return statuses.length === 0 || statuses.includes(authStatusEnum(mailbox));
}

export function bulkMailboxActionCount(mailboxes: Mailbox[], action: MailboxProviderActionCapability | undefined) {
  if (!action?.bulk_supported) return 0;
  return mailboxes.filter((mailbox) => canRunMailboxAction(mailbox, action)).length;
}

export function canRunProviderMailboxAction(capabilities: MailboxProviderCapability[], mailbox: Mailbox, action: MailboxProviderAction) {
  return canRunMailboxAction(mailbox, providerAction(capabilityForProvider(capabilities, mailbox.provider_key), action));
}

export function providerShowsCredentialState(capability?: MailboxProviderCapability) {
  return (capability?.actions || []).some((action) => (action.required_credentials || []).length > 0);
}

export function providerDisplayName(capability: MailboxProviderCapability | undefined, fallback: string) {
  return capability?.display_name || fallback;
}

function requiredCredentialsPresent(mailbox: Mailbox, credentials: MailboxCredentialKind[]) {
  return credentials.every((credential) => mailboxCredentialPresent(mailbox, credential));
}

export function mailboxCredentialPresent(mailbox: Mailbox, credential: MailboxCredentialKind) {
  switch (credential) {
    case MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_UNSPECIFIED:
      return true;
    default:
      return (mailbox.credential_state?.present_credentials || []).includes(credential);
  }
}

function normalizeAuthStatus(value: string) {
  return authStatusNameByValue.get(value) || '';
}
