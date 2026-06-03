import { MailboxAuthStatus, MailboxCredentialKind, MailboxProviderAction } from '@byte-v-forge/common-ui';
import { normalizeUiEmail } from './email-utils';
import { mailboxProviderMatches, mailboxProviderValue, normalizeMailboxProviderKey, type MailboxProviderTab } from './mailbox-provider-config';
import type { Mailbox, MailboxProviderActionCapability, MailboxProviderCapability } from './types';

export { mailboxProviderMatches, mailboxProviderValue, normalizeMailboxProviderKey, type MailboxProviderTab };
export type MailboxBatchItem = {
  email: string;
  password: string;
};

export function domainForEmail(email: string) {
  const [, domain = ''] = normalizeUiEmail(email).split('@');
  return domain;
}

export function tokenText(mailbox: Mailbox) {
  if (mailbox.refresh_token && authStatus(mailbox) === 'AUTHORIZED') return 'Refresh 可用';
  if (mailbox.refresh_token) return 'Refresh 待验证';
  if (mailbox.access_token) return '仅 Access';
  return '缺 Token';
}

export function authStatus(mailbox: Mailbox) {
  const value = normalizeAuthStatus(String(mailbox.auth_status || '').trim());
  if (value) return value;
  if (mailbox.refresh_token) return 'AUTHORIZED';
  return 'OAUTH_PENDING';
}

export function authStatusEnum(mailbox: Mailbox) {
  switch (authStatus(mailbox)) {
    case 'AUTHORIZED':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_AUTHORIZED;
    case 'OAUTH_PENDING':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_OAUTH_PENDING;
    case 'AUTH_FAILED':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_AUTH_FAILED;
    case 'NEEDS_MANUAL_VERIFICATION':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_NEEDS_MANUAL_VERIFICATION;
    case 'PASSWORD_ONLY':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_PASSWORD_ONLY;
    case 'WEBHOOK_ONLY':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_WEBHOOK_ONLY;
    case 'DISABLED':
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_DISABLED;
    default:
      return MailboxAuthStatus.MAILBOX_AUTH_STATUS_UNKNOWN;
  }
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
  return credentials.every((credential) => credentialPresent(mailbox, credential));
}

function credentialPresent(mailbox: Mailbox, credential: MailboxCredentialKind) {
  switch (credential) {
    case MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_UNSPECIFIED:
      return true;
    case MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD:
      return !!String(mailbox.password || '').trim();
    case MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN:
      return !!String(mailbox.refresh_token || '').trim();
    case MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN:
      return !!String(mailbox.access_token || '').trim();
    default:
      return false;
  }
}

function normalizeAuthStatus(value: string) {
  switch (value) {
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_AUTHORIZED:
    case 'AUTHORIZED':
      return 'AUTHORIZED';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_OAUTH_PENDING:
    case 'OAUTH_PENDING':
      return 'OAUTH_PENDING';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_AUTH_FAILED:
    case 'AUTH_FAILED':
      return 'AUTH_FAILED';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_NEEDS_MANUAL_VERIFICATION:
    case 'NEEDS_MANUAL_VERIFICATION':
      return 'NEEDS_MANUAL_VERIFICATION';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_PASSWORD_ONLY:
    case 'PASSWORD_ONLY':
      return 'PASSWORD_ONLY';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_WEBHOOK_ONLY:
    case 'WEBHOOK_ONLY':
      return 'WEBHOOK_ONLY';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_DISABLED:
    case 'DISABLED':
      return 'DISABLED';
    case MailboxAuthStatus.MAILBOX_AUTH_STATUS_UNKNOWN:
    case 'UNKNOWN':
      return 'UNKNOWN';
    default:
      return '';
  }
}
