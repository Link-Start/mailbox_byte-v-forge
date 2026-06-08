import { useMemo } from 'react';
import { useSearchParams } from 'react-router';
import { normalizeUiEmail } from './email-utils';
import {
  activeMailboxPanelProvider,
  mailboxPanelImportProvider,
  mailboxPanelProvider,
  mailboxPanelQuery,
  nextMailboxPanelParams,
  type MailboxPanelQueryUpdate
} from './mailbox-panel-query';
import { mailboxProviderViews } from './mailbox-provider-registry';
import type { Mailbox, MailboxProviderCapability } from './types';

export function useMailboxPanelState(mailboxes: Mailbox[], providerCapabilities: MailboxProviderCapability[]) {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = mailboxPanelQuery(searchParams);
  const searchQuery = query.trim();
  const filteredMailboxes = useMemo(() => filterMailboxes(mailboxes, searchQuery), [mailboxes, searchQuery]);
  const providerViews = useMemo(() => mailboxProviderViews(providerCapabilities, filteredMailboxes), [providerCapabilities, filteredMailboxes]);

  function updateQuery(update: MailboxPanelQueryUpdate) {
    setSearchParams((current) => nextMailboxPanelParams(current, update), { replace: true });
  }

  return {
    query,
    searchQuery,
    filteredMailboxes,
    providerViews,
    activeProvider: activeMailboxPanelProvider(providerViews, mailboxPanelProvider(searchParams)),
    importProvider: mailboxPanelImportProvider(searchParams),
    updateQuery
  };
}

function filterMailboxes(mailboxes: Mailbox[], query: string) {
  const needle = query.trim().toLowerCase();
  if (!needle) return mailboxes;
  return mailboxes.filter((mailbox) => [
    mailbox.email_address,
    mailbox.domain,
    mailbox.provider_key
  ].some((value) => normalizeUiEmail(value || '').includes(needle)));
}
