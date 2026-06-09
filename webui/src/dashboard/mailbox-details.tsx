import { MailboxInboxSection } from './mailbox-inbox';
import { MailboxOverview } from './mailbox-overview';
import type { MailboxPanelMode } from './mailbox-provider-types';
import type { Mailbox, MailboxProviderCapability } from './types';

export function MailboxDetails({ mailbox, providerCapability, mode, inboxLoading, canFetchInbox, onCopy, onFetchInbox, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  mode: MailboxPanelMode;
  inboxLoading: boolean;
  canFetchInbox: boolean;
  onCopy: (label: string, value: string) => void;
  onFetchInbox: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  if (mode === 'inbox') {
    return <MailboxInboxSection mailbox={mailbox} loading={inboxLoading} canFetch={canFetchInbox} onFetch={onFetchInbox} />;
  }
  return <MailboxOverview mailbox={mailbox} providerCapability={providerCapability} onCopy={onCopy} onDelete={onDelete} />;
}
