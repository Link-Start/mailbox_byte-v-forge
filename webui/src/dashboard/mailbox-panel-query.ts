import { normalizeMailboxProviderKey, type MailboxProviderTab } from './mailbox-provider-config';

type ProviderView = { value: string };

type MailboxPanelQueryUpdate = {
  provider?: string;
  q?: string;
  importProvider?: string;
};

const providerParam = 'provider';
const queryParam = 'q';
const importProviderParam = 'import';

export function mailboxPanelProvider(params: URLSearchParams): MailboxProviderTab {
  return normalizeMailboxProviderKey(params.get(providerParam) || '');
}

export function mailboxPanelQuery(params: URLSearchParams) {
  return (params.get(queryParam) || '').trim();
}

export function mailboxPanelImportProvider(params: URLSearchParams): MailboxProviderTab {
  return normalizeMailboxProviderKey(params.get(importProviderParam) || '');
}

export function activeMailboxPanelProvider(views: ProviderView[], requested: MailboxProviderTab): MailboxProviderTab {
  if (views.some((view) => view.value === requested)) return requested;
  return views[0]?.value || '';
}

export function nextMailboxPanelParams(current: URLSearchParams, update: MailboxPanelQueryUpdate) {
  const next = new URLSearchParams(current);
  if ('provider' in update) setParam(next, providerParam, normalizeMailboxProviderKey(update.provider || ''));
  if ('q' in update) setParam(next, queryParam, update.q?.trim());
  if ('importProvider' in update) setParam(next, importProviderParam, normalizeMailboxProviderKey(update.importProvider || ''));
  return next;
}

export function persistentMailboxPanelSearch(search: string) {
  const params = new URLSearchParams(search);
  params.delete(importProviderParam);
  const value = params.toString();
  return value ? `?${value}` : '';
}

function setParam(params: URLSearchParams, key: string, value?: string) {
  if (value === undefined) return;
  if (value) params.set(key, value);
  else params.delete(key);
}
