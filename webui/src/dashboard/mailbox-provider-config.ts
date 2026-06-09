export type MailboxProviderTab = string;
export const mailboxAllProviderTab = 'all';

export function mailboxProviderValue(provider: string): string {
  return normalizeMailboxProviderKey(provider);
}

export function mailboxProviderMatches(provider: string, target: string) {
  return normalizeMailboxProviderKey(provider) === normalizeMailboxProviderKey(target);
}

export function normalizeMailboxProviderKey(provider: string) {
  return String(provider || '').trim().toLowerCase();
}
