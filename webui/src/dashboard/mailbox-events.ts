import type { QueryKey } from '@tanstack/react-query';
import { createHotStreamURL, useHotStreamInvalidation } from './hotstream';
import { normalizeUiEmail } from './email-utils';
import type { InboxMessage, InboxResult, Mailbox } from './types';

const mailboxApiBase = '/api/mailbox';

export type MailboxEmailEventCacheOptions = {
  enabled?: boolean;
  email?: string;
  signalKind?: 'otp' | 'any';
  issuedAfterUnix?: number;
  inboxQueryKey?: QueryKey;
};

export function useMailboxEmailEventCache(options: MailboxEmailEventCacheOptions) {
  const email = normalizeUiEmail(options.email || '');
  useHotStreamInvalidation({
    enabled: options.enabled !== false && !!email && !!options.inboxQueryKey,
    url: mailboxEventURL(email, options.signalKind, options.issuedAfterUnix),
    rules: options.inboxQueryKey ? [{ queryKey: options.inboxQueryKey, eventTypes: eventTypes(options.signalKind), resourceTypes: ['mailbox.email'], resourceIds: [email] }] : []
  });
}

export function mergeInboxMessage(result: InboxResult | null | undefined, email: string, message: InboxMessage): InboxResult {
  const target = normalizeUiEmail(email || message.mailbox_email);
  const current: InboxResult = result ? { ...result, messages: [...(result.messages || [])] } : { mailbox: { email_address: target } as Mailbox, messages: [], error_message: '' };
  current.messages = [message, ...(current.messages || []).filter((item) => messageKey(item) !== messageKey(message))]
    .sort((a, b) => (b.received_at_unix || 0) - (a.received_at_unix || 0))
    .slice(0, 20);
  return current;
}

export function mailboxEventURL(email: string, signalKind = 'otp', _issuedAfterUnix = 0) {
  return createHotStreamURL(mailboxApiBase, { eventTypes: eventTypes(signalKind), resourceTypes: ['mailbox.email'], resourceIds: [email] });
}

function eventTypes(signalKind = 'otp') {
  return signalKind === 'any' ? ['mailbox.email.received', 'mailbox.email.signal_received'] : ['mailbox.email.signal_received'];
}

function messageKey(message: InboxMessage) {
  return [message.provider_key || '', message.mailbox_email || '', message.id || `${message.received_at_unix}:${message.subject}`].join(':');
}
