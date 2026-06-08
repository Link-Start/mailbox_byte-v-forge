import {
  api,
  type ListMailboxDomainsResponse,
  type ListMailboxOperationsResponse,
  type ListMailboxProviderCapabilitiesResponse
} from './dashboard-kit';
import { mailboxApiPaths, mailboxListURL, mailboxLookupURL, mailboxRunningOperationsURL } from './mailbox-api-paths';
import type { ListEmailMailboxesResponse } from './types';

export function listMailboxes(cursor: string) {
  return api<ListEmailMailboxesResponse>(mailboxListURL(cursor));
}

export function lookupMailbox(email: string) {
  return api<ListEmailMailboxesResponse>(mailboxLookupURL(email));
}

export function listMailboxDomains() {
  return api<ListMailboxDomainsResponse>(mailboxApiPaths.domains);
}

export function listMailboxProviderCapabilities() {
  return api<ListMailboxProviderCapabilitiesResponse>(mailboxApiPaths.providerCapabilities);
}

export function listRunningMailboxOperations() {
  return api<ListMailboxOperationsResponse>(mailboxRunningOperationsURL());
}
