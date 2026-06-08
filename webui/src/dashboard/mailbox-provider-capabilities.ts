import { MailboxProviderAction } from './dashboard-kit';
import { authStatusEnum } from './mailbox-auth-status';
import { requiredCredentialsPresent } from './mailbox-credentials';
import { mailboxProviderMatches, normalizeMailboxProviderKey } from './mailbox-provider-config';
import type { Mailbox, MailboxProviderActionCapability, MailboxProviderCapability } from './types';

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
