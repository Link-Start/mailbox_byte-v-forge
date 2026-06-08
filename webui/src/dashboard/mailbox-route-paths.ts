export const mailboxStandaloneBasename = '/dashboard/mailbox';
export const mailboxDashboardViewBasename = '/mailbox/mailboxes';

export type MailboxDetailTab = 'overview' | 'inbox';

export function mailboxIndexPath() {
  return '/';
}

export function mailboxDetailPath(email: string, tab: MailboxDetailTab = 'overview') {
  const base = `/mailboxes/${encodeURIComponent(email)}`;
  return tab === 'inbox' ? `${base}/inbox` : base;
}
