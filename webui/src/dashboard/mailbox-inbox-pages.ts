import { useMemo } from 'react';
import { useInfiniteQuery, type InfiniteData, type QueryKey } from '@tanstack/react-query';
import { normalizeUiEmail } from './email-utils';
import { fetchStoredInboxPage } from './mailbox-action-api';
import type { InboxMessage, ListMailboxInboxResponse } from './types';

export const mailboxInboxPageSize = 20;

export const mailboxInboxPagesQueryKey = (email: string, query: string) => [
  'mailbox',
  'inbox',
  normalizeUiEmail(email),
  'pages',
  query.trim()
] as const;

export function useMailboxInboxPages(email: string, query: string) {
  const target = normalizeUiEmail(email);
  const search = query.trim();
  const inboxQuery = useInfiniteQuery<ListMailboxInboxResponse, Error, InfiniteData<ListMailboxInboxResponse>, QueryKey, string>({
    queryKey: mailboxInboxPagesQueryKey(target, search),
    queryFn: ({ pageParam }) => fetchStoredInboxPage(target, { cursor: pageParam, limit: mailboxInboxPageSize, query: search }),
    initialPageParam: '',
    getNextPageParam: (page) => page.next_cursor || undefined,
    enabled: !!target,
  });
  const messages = useMemo(() => (inboxQuery.data?.pages || []).flatMap((page) => page.result?.messages || []), [inboxQuery.data?.pages]);
  const errorPage = (inboxQuery.data?.pages || []).find((page) => page.error_message || page.result?.error_message);
  const errorMessage = errorPage?.error_message || errorPage?.result?.error_message || inboxQuery.error?.message || '';
  return {
    ...inboxQuery,
    messages: messages as InboxMessage[],
    errorMessage,
    loadMore: () => inboxQuery.fetchNextPage().then(() => undefined),
  };
}
