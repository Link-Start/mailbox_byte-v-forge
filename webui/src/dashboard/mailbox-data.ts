import { useMemo } from 'react';
import {
  DEFAULT_CURSOR_PAGE_SIZE,
  createHotStreamURL,
  useHotStreamInvalidation,
  useCursorPageItems,
  useQuery,
  useQueryClient
} from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { mailboxApiPaths } from './mailbox-api-paths';
import { listMailboxProviderCapabilities, listMailboxes, listRunningMailboxOperations, lookupMailbox } from './mailbox-data-api';
import type { ListEmailMailboxesResponse, Mailbox, MailboxOperation } from './types';

const mailboxQueryKeys = {
  mailboxes: ['mailbox', 'mailboxes'] as const,
  mailbox: (email: string) => ['mailbox', 'mailbox', normalizeUiEmail(email)] as const,
  providerCapabilities: ['mailbox', 'provider-capabilities'] as const,
  runningOperations: ['mailbox', 'running-operations'] as const
};

export function useMailboxData(selectedEmail: string) {
  const queryClient = useQueryClient();
  const mailboxesQuery = useCursorPageItems<Mailbox, ListEmailMailboxesResponse, 'mailboxes'>({
    queryKey: mailboxQueryKeys.mailboxes,
    field: 'mailboxes',
    queryFn: listMailboxes,
    pageSize: DEFAULT_CURSOR_PAGE_SIZE
  });
  const providerCapabilitiesQuery = useQuery({ queryKey: mailboxQueryKeys.providerCapabilities, queryFn: listMailboxProviderCapabilities });
  const runningOperationsQuery = useQuery({
    queryKey: mailboxQueryKeys.runningOperations,
    queryFn: listRunningMailboxOperations
  });
  useHotStreamInvalidation({
    url: createHotStreamURL(mailboxApiPaths.base, { eventTypes: ['mailbox.email.received', 'mailbox.email.signal_received', 'mailbox.operation.updated'] }),
    rules: [
      { queryKey: mailboxQueryKeys.mailboxes, eventTypes: ['mailbox.email.received', 'mailbox.email.signal_received'], resourceTypes: ['mailbox.email'] },
      { queryKey: mailboxQueryKeys.runningOperations, eventTypes: ['mailbox.operation.updated'], resourceTypes: ['mailbox.operation'] }
    ]
  });
  const mailboxes = mailboxesQuery.items;
  const selectedFromList = mailboxes.find((mailbox) => mailbox.email_address === selectedEmail) || null;
  const selectedQuery = useQuery({
    queryKey: mailboxQueryKeys.mailbox(selectedEmail),
    queryFn: () => lookupMailbox(selectedEmail),
    enabled: !!selectedEmail && !selectedFromList
  });
  const runningOperations = runningOperationsQuery.data?.operations;
  const selected = selectedFromList || selectedQuery.data?.mailboxes?.[0] || null;
  const runningOperationByEmail = useMemo(() => latestOperationByEmail(Array.isArray(runningOperations) ? runningOperations : []), [runningOperations]);

  return {
    mailboxes,
    selected,
    runningOperationByEmail,
    providerCapabilities: Array.isArray(providerCapabilitiesQuery.data?.providers) ? providerCapabilitiesQuery.data.providers : [],
    busy: mailboxesQuery.isLoading || providerCapabilitiesQuery.isLoading || selectedQuery.isLoading,
    hasMoreMailboxes: mailboxesQuery.pagination.hasNext,
    loadingMoreMailboxes: mailboxesQuery.pagination.loading,
    loadMoreMailboxes: mailboxesQuery.loadMore,
    loadError: mailboxesQuery.error || providerCapabilitiesQuery.error || runningOperationsQuery.error,
    invalidate: () => invalidateMailboxQueries(queryClient)
  };
}

export type MailboxData = ReturnType<typeof useMailboxData>;

function invalidateMailboxQueries(queryClient: ReturnType<typeof useQueryClient>) {
  return queryClient.invalidateQueries({ queryKey: ['mailbox'] });
}

function latestOperationByEmail(operations: MailboxOperation[]) {
  const out = new Map<string, MailboxOperation>();
  for (const operation of operations) {
    const email = normalizeUiEmail(operation.email_address);
    if (!email) continue;
    const previous = out.get(email);
    if (!previous || (operation.updated_at || 0) > (previous.updated_at || 0)) out.set(email, operation);
  }
  return out;
}
