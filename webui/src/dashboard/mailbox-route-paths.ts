export const mailboxRouterBasename = '/dashboard/mailbox';

export type MailboxDetailTab = 'overview' | 'inbox';

export function mailboxIndexPath() {
  return '/';
}

export function mailboxDetailPath(email: string, tab: MailboxDetailTab = 'overview') {
  const base = `/mailboxes/${encodeURIComponent(email)}`;
  return tab === 'inbox' ? `${base}/inbox` : base;
}
