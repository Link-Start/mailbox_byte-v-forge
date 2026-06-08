import { api } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { mailboxInboxMessageURL } from './mailbox-api-paths';
import type { GetMailboxInboxMessageRequest, GetMailboxInboxMessageResponse } from './contracts';
import type { InboxMessage } from './types';

export const mailboxInboxMessageQueryKey = (email: string, messageID: string, providerKey: string) => [
  'mailbox',
  'inbox-message',
  normalizeUiEmail(email),
  messageID.trim(),
  providerKey.trim().toLowerCase()
] as const;

export async function fetchInboxMessageDetail(message: InboxMessage): Promise<GetMailboxInboxMessageResponse | null> {
  const input = inboxMessageDetailInput(message);
  if (!input.email_address || !input.message_id) return null;
  const query = new URLSearchParams();
  if (input.provider_key) query.set('provider_key', input.provider_key);
  if (input.parser_profile) query.set('parser_profile', input.parser_profile);
  const queryString = query.toString();
  return api<GetMailboxInboxMessageResponse>(mailboxInboxMessageURL(input.email_address, input.message_id, queryString));
}

function inboxMessageDetailInput(message: InboxMessage): GetMailboxInboxMessageRequest {
  return {
    email_address: normalizeUiEmail(message.mailbox_email),
    message_id: message.id || '',
    provider_key: message.provider_key || '',
    parser_profile: ''
  };
}
