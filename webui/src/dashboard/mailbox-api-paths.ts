import { DEFAULT_CURSOR_PAGE_SIZE, cursorPageURL } from './dashboard-kit';

export const mailboxApiBase = '/api/mailbox';

export const mailboxApiPaths = {
  base: mailboxApiBase,
  domains: `${mailboxApiBase}/domains`,
  mailboxes: `${mailboxApiBase}/mailboxes`,
  mailboxOAuth: `${mailboxApiBase}/mailboxes/oauth`,
  mailboxInboxFetch: `${mailboxApiBase}/mailboxes/inbox`,
  operations: `${mailboxApiBase}/operations`,
  providerCapabilities: `${mailboxApiBase}/provider-capabilities`
} as const;

export function mailboxURL(email: string) {
  return `${mailboxApiPaths.mailboxes}/${encodeURIComponent(email)}`;
}

export function mailboxInboxURL(email: string, limit = 20) {
  return `${mailboxURL(email)}/inbox?limit=${limit}`;
}

export function mailboxInboxMessageURL(email: string, messageID: string, queryString = '') {
  return `${mailboxURL(email)}/inbox/${encodeURIComponent(messageID)}${queryString ? `?${queryString}` : ''}`;
}

export function mailboxListURL(cursor: string) {
  return cursorPageURL(mailboxApiPaths.mailboxes, { cursor, limit: DEFAULT_CURSOR_PAGE_SIZE });
}

export function mailboxLookupURL(email: string) {
  return cursorPageURL(mailboxApiPaths.mailboxes, { cursor: '', limit: 1, params: { email_address: email } });
}
