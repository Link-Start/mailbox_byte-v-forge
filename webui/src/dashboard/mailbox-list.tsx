import { DEFAULT_CURSOR_PAGE_SIZE, CursorPager, EmptyBlock, ScrollArea } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { MailboxCard } from './mailbox-card';
import type { MailboxPanelMode } from './mailbox-provider-types';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';

export type MailboxRecordListProps = {
  mailboxes: Mailbox[];
  mode: MailboxPanelMode;
  emptyText: string;
  providerCapability?: MailboxProviderCapability;
  providerCapabilities?: MailboxProviderCapability[];
  showStatus?: boolean;
  selected?: string;
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  runningOperationByEmail: Map<string, MailboxOperation>;
  hasMoreMailboxes?: boolean;
  loadingMoreMailboxes?: boolean;
  onLoadMoreMailboxes?: () => void | Promise<void>;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
};

export function MailboxRecordList({ mailboxes, mode, emptyText, providerCapability, providerCapabilities, showStatus, selected, busy, showSecrets, oauthing, runningOperationByEmail, hasMoreMailboxes, loadingMoreMailboxes, onLoadMoreMailboxes, onOAuth, onDelete }: MailboxRecordListProps) {
  return (
    <>
      <ScrollArea className="h-full min-h-0">
        <div className="recordList wideRecordList">
          {mailboxes.length > 0 ? mailboxes.map((mailbox) => (
            <MailboxCard
              key={mailbox.email_address}
              mailbox={mailbox}
              mode={mode}
              selected={selected === mailbox.email_address}
              busy={busy}
              showSecrets={showSecrets}
              oauthing={oauthing}
              showStatus={showStatus ?? true}
              providerCapability={providerCapability}
              providerCapabilities={providerCapabilities}
              currentOperation={runningOperationByEmail.get(normalizeUiEmail(mailbox.email_address))}
              onOAuth={onOAuth}
              onDelete={onDelete}
            />
          )) : <EmptyBlock text={emptyText} />}
        </div>
      </ScrollArea>
      {onLoadMoreMailboxes && <CursorPager itemCount={mailboxes.length} pageSize={DEFAULT_CURSOR_PAGE_SIZE} hasNext={hasMoreMailboxes} loading={loadingMoreMailboxes} onNext={() => void onLoadMoreMailboxes()} />}
    </>
  );
}
