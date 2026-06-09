import { mailboxAllProviderTab, mailboxProviderMatches, normalizeMailboxProviderKey } from './mailbox-provider-config';
import type { Mailbox, MailboxProviderCapability } from './types';

export type MailboxProviderView = {
  value: string;
  label: string;
  capability?: MailboxProviderCapability;
  mailboxes: Mailbox[];
};

export function mailboxProviderViews(capabilities: MailboxProviderCapability[], mailboxes: Mailbox[]): MailboxProviderView[] {
  const seen = new Set<string>([mailboxAllProviderTab]);
  const views = [allMailboxView(mailboxes), ...capabilities
    .map((capability) => providerView(capability.key, capability, mailboxes))
    .filter((view): view is MailboxProviderView => !!view && markSeen(seen, view.value))];
  for (const mailbox of mailboxes) {
    const key = normalizeMailboxProviderKey(mailbox.provider_key);
    const view = providerView(key, undefined, mailboxes);
    if (view && markSeen(seen, view.value)) views.push(view);
  }
  return views;
}

function allMailboxView(mailboxes: Mailbox[]): MailboxProviderView {
  return {
    value: mailboxAllProviderTab,
    label: '全部',
    mailboxes,
  };
}

function providerView(providerKey: string, capability: MailboxProviderCapability | undefined, mailboxes: Mailbox[]): MailboxProviderView | null {
  const value = normalizeMailboxProviderKey(providerKey);
  if (!value) return null;
  return {
    value,
    label: capability?.display_name || value,
    capability,
    mailboxes: mailboxes.filter((mailbox) => mailboxProviderMatches(mailbox.provider_key, value)),
  };
}

function markSeen(seen: Set<string>, value: string) {
  if (seen.has(value)) return false;
  seen.add(value);
  return true;
}
