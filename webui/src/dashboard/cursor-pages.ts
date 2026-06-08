import { useInfiniteQuery, type InfiniteData, type QueryKey } from '@tanstack/react-query';

export const DEFAULT_CURSOR_PAGE_SIZE = 100;

type CursorPageResponse = { next_cursor?: string | null };

type CursorPageOptions<R extends object> = {
  queryKey: QueryKey;
  queryFn: (cursor: string) => Promise<R>;
  field: keyof R;
  enabled?: boolean;
  initialCursor?: string;
  pageSize?: number;
};

export function useCursorPageItems<T, R extends object, K extends keyof R>(options: CursorPageOptions<R> & { field: K }) {
  const query = useInfiniteQuery<R, Error, InfiniteData<R>, QueryKey, string>({
    queryKey: options.queryKey,
    queryFn: ({ pageParam }) => options.queryFn(pageParam || ''),
    initialPageParam: options.initialCursor || '',
    getNextPageParam: (lastPage) => cursorPageNextCursor(lastPage as CursorPageResponse) || undefined,
    enabled: options.enabled,
  });
  return {
    ...query,
    items: (query.data?.pages || []).flatMap((page) => responseList<T, R, K>(page, options.field)),
    loadMore: () => query.fetchNextPage().then(() => undefined),
    pagination: {
      pageSize: options.pageSize ?? DEFAULT_CURSOR_PAGE_SIZE,
      hasNext: query.hasNextPage,
      loading: query.isFetchingNextPage,
    },
  };
}

export function cursorPageURL(path: string, options: { cursor?: string; limit?: number; params?: Record<string, string | number | boolean | null | undefined> }) {
  const params = new URLSearchParams();
  params.set('limit', String(options.limit ?? DEFAULT_CURSOR_PAGE_SIZE));
  for (const [key, value] of Object.entries(options.params || {})) {
    const text = String(value ?? '').trim();
    if (text) params.set(key, text);
  }
  const cursor = (options.cursor || '').trim();
  if (cursor) params.set('cursor', cursor);
  return `${path}?${params.toString()}`;
}

function cursorPageNextCursor(response: CursorPageResponse | null | undefined) {
  return (response?.next_cursor || '').trim();
}

function responseList<T, R extends object, K extends keyof R>(response: R | null | undefined, field: K): T[] {
  const value = response?.[field];
  return Array.isArray(value) ? (value as T[]) : [];
}
