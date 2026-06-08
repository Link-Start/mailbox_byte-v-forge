import {
  api,
  type FetchMailboxInboxesRequest,
  type ListMailboxInboxResponse,
  type StartMailboxOAuthRequest,
  type StartMailboxOAuthResponse,
  type SyncMailboxDomainsRequest,
  type SyncMailboxDomainsResponse
} from './dashboard-kit';
import { mailboxApiPaths, mailboxInboxURL, mailboxURL } from './mailbox-api-paths';
import type { DeleteMailboxResponse, InboxResponse } from './types';

export function startMailboxOAuth(emailAddress = '') {
  const input: StartMailboxOAuthRequest = {
    email_address: emailAddress,
    only_missing: !emailAddress,
    limit: 100
  };
  return api<StartMailboxOAuthResponse>(mailboxApiPaths.mailboxOAuth, { method: 'POST', body: JSON.stringify(input) });
}

export function fetchMailboxInboxes(emailAddress: string) {
  const input: FetchMailboxInboxesRequest = {
    limit_per_mailbox: 10,
    max_mailboxes: emailAddress ? 1 : 200,
    email_address: emailAddress,
    parser_profile: '',
    received_after_unix: 0
  };
  return api<InboxResponse>(mailboxApiPaths.mailboxInboxFetch, { method: 'POST', body: JSON.stringify(input) });
}

export function syncMailboxDomains(providerKey: string) {
  const input: SyncMailboxDomainsRequest = { provider_key: providerKey };
  return api<SyncMailboxDomainsResponse>(mailboxApiPaths.domains, { method: 'POST', body: JSON.stringify(input) });
}

export function deleteMailbox(email: string) {
  return api<DeleteMailboxResponse>(mailboxURL(email), { method: 'DELETE' });
}

export async function fetchStoredInbox(email: string) {
  const resp = await api<ListMailboxInboxResponse>(mailboxInboxURL(email));
  return resp.result || null;
}
