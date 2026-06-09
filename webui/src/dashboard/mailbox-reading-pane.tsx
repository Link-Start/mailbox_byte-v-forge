import { Mail, X } from 'lucide-react';
import { Button, EmptyBlock } from './dashboard-kit';
import { maskEmail } from './email-utils';
import { MailboxDetails } from './mailbox-details';
import type { MailboxDetailTab } from './mailbox-route-paths';
import type { InboxResult, Mailbox, MailboxProviderCapability } from './types';

export function MailboxReadingEmpty({ busy, total }: { busy: boolean; total: number }) {
  const text = busy ? '加载中' : total > 0 ? '选择邮箱' : '暂无邮箱';
  return <div className="mailboxReadingEmpty"><EmptyBlock text={text} /></div>;
}

export function MailboxReadingPane({
  mailbox,
  providerCapability,
  activeTab,
  showSecrets,
  inboxResult,
  inboxLoading,
  canFetchInbox,
  onClose,
  onTabChange,
  onCopy,
  onFetchInbox,
  onDelete
}: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  activeTab: MailboxDetailTab;
  showSecrets: boolean;
  inboxResult?: InboxResult | null;
  inboxLoading: boolean;
  canFetchInbox: boolean;
  onClose: () => void;
  onTabChange: (tab: MailboxDetailTab) => void;
  onCopy: (label: string, value: string) => void;
  onFetchInbox: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const title = showSecrets ? mailbox.email_address : maskEmail(mailbox.email_address);
  return (
    <section className="mailboxDetailPane">
      <header className="mailboxDetailHeader">
        <div className="recordIdentity min-w-0">
          <span className="recordIcon"><Mail className="size-4" /></span>
          <strong className="recordTitle" title={title}>{title}</strong>
        </div>
        <Button variant="ghost" size="icon" title="关闭" aria-label="关闭" onClick={onClose}>
          <X className="size-4" />
        </Button>
      </header>
      <div className="mailboxDetailBody">
        <MailboxDetails
          mailbox={mailbox}
          providerCapability={providerCapability}
          activeTab={activeTab}
          showSecrets={showSecrets}
          inboxResult={inboxResult}
          inboxLoading={inboxLoading}
          canFetchInbox={canFetchInbox}
          onTabChange={onTabChange}
          onCopy={onCopy}
          onFetchInbox={onFetchInbox}
          onDelete={onDelete}
        />
      </div>
    </section>
  );
}
