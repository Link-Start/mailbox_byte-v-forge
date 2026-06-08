import { ContentTabs } from './dashboard-kit';
import { MailboxInboxSection } from './mailbox-inbox';
import { MailboxOverview } from './mailbox-overview';
import type { MailboxDetailTab } from './mailbox-route-paths';
import { latestOtpForInboxResult } from './mailbox-signal-utils';
import type { InboxResult, Mailbox, MailboxProviderCapability } from './types';

export function MailboxDetails({ mailbox, providerCapability, activeTab, showSecrets, inboxResult, inboxLoading, canFetchInbox, onTabChange, onCopy, onFetchInbox, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  activeTab: MailboxDetailTab;
  showSecrets: boolean;
  inboxResult?: InboxResult | null;
  inboxLoading: boolean;
  canFetchInbox: boolean;
  onTabChange: (tab: MailboxDetailTab) => void;
  onCopy: (label: string, value: string) => void;
  onFetchInbox: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const inboxMessageCount = inboxResult?.messages?.length || 0;
  const latestOtp = latestOtpForInboxResult(inboxResult || null, mailbox.email_address);

  return (
    <ContentTabs
      value={activeTab}
      onValueChange={(value) => onTabChange(value as MailboxDetailTab)}
      tabsClassName="min-h-0 flex-1 overflow-hidden"
      tabsListVariant="line"
      tabsListClassName="w-full"
      tabs={[
        {
          value: 'overview',
          label: '概览',
          contentClassName: 'overflow-auto',
          content: <MailboxOverview mailbox={mailbox} providerCapability={providerCapability} showSecrets={showSecrets} latestOtp={latestOtp} onCopy={onCopy} onDelete={onDelete} />
        },
        {
          value: 'inbox',
          label: `收件箱 ${inboxMessageCount}`,
          contentClassName: 'overflow-auto',
          content: <MailboxInboxSection mailbox={mailbox} result={inboxResult} showSecrets={showSecrets} loading={inboxLoading} canFetch={canFetchInbox} onFetch={onFetchInbox} />
        }
      ]}
    />
  );
}
