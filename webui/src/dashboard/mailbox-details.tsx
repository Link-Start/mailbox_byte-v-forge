import { MailboxInboxSection } from './mailbox-inbox';
import { MailboxOverview } from './mailbox-overview';
import type { MailboxPanelMode } from './mailbox-provider-types';
import { latestOtpForInboxResult } from './mailbox-signal-utils';
import type { InboxResult, Mailbox, MailboxProviderCapability } from './types';

export function MailboxDetails({ mailbox, providerCapability, mode, showSecrets, inboxResult, inboxLoading, canFetchInbox, onCopy, onFetchInbox, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  mode: MailboxPanelMode;
  showSecrets: boolean;
  inboxResult?: InboxResult | null;
  inboxLoading: boolean;
  canFetchInbox: boolean;
  onCopy: (label: string, value: string) => void;
  onFetchInbox: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  if (mode === 'inbox') {
    return <MailboxInboxSection mailbox={mailbox} result={inboxResult} showSecrets={showSecrets} loading={inboxLoading} canFetch={canFetchInbox} onFetch={onFetchInbox} />;
  }
  const latestOtp = latestOtpForInboxResult(inboxResult || null, mailbox.email_address);
  return <MailboxOverview mailbox={mailbox} providerCapability={providerCapability} showSecrets={showSecrets} latestOtp={latestOtp} onCopy={onCopy} onDelete={onDelete} />;
}
