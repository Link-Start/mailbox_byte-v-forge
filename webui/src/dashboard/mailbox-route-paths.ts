export const mailboxStandaloneBasename = '/dashboard/mailbox';
export const mailboxDashboardViewBasename = '/mailbox/mailboxes';

export type MailboxDetailTab = 'overview' | 'inbox';

export function mailboxIndexPath() {
  return '/';
}

export function mailboxInboxIndexPath() {
  return '/inbox';
}

export function mailboxAccountsIndexPath() {
  return '/accounts';
}

export function mailboxDetailPath(email: string, tab: MailboxDetailTab = 'overview') {
  const encoded = encodeURIComponent(email);
  return tab === 'inbox' ? `/inbox/${encoded}` : `/accounts/${encoded}`;
}
