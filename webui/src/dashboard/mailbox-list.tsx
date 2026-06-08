import { DEFAULT_CURSOR_PAGE_SIZE, CursorPager, RecordList } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { uniqueStrings } from './dashboard-kit';
import { domainForEmail } from './mailbox-email';
import { MailboxCard } from './mailbox-card';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';

export function MailboxRecordList({ mailboxes, emptyText, providerCapability, showStatus, selected, busy, showSecrets, oauthing, runningOperationByEmail, hasMoreMailboxes, loadingMoreMailboxes, onLoadMoreMailboxes, onOAuth, onDelete }: {
  mailboxes: Mailbox[];
  emptyText: string;
  providerCapability?: MailboxProviderCapability;
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
}) {
  return (
    <>
      <RecordList className="wideRecordList" emptyText={emptyText}>
        {mailboxes.map((mailbox) => (
          <MailboxCard
            key={mailbox.email_address}
            mailbox={mailbox}
            selected={selected === mailbox.email_address}
            busy={busy}
            showSecrets={showSecrets}
            oauthing={oauthing}
            showStatus={showStatus ?? true}
            providerCapability={providerCapability}
            currentOperation={runningOperationByEmail.get(normalizeUiEmail(mailbox.email_address))}
            onOAuth={onOAuth}
            onDelete={onDelete}
          />
        ))}
      </RecordList>
      {onLoadMoreMailboxes && <CursorPager itemCount={mailboxes.length} pageSize={DEFAULT_CURSOR_PAGE_SIZE} hasNext={hasMoreMailboxes} loading={loadingMoreMailboxes} onNext={() => void onLoadMoreMailboxes()} />}
    </>
  );
}

export function MailboxDomainGroups(props: {
  mailboxes: Mailbox[];
  configuredDomains: string[];
  providerCapability?: MailboxProviderCapability;
  showStatus?: boolean;
  emptyDomainsText: string;
  emptyDomainText: string;
  selected?: string;
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  runningOperationByEmail: Map<string, MailboxOperation>;
  hasMoreMailboxes?: boolean;
  loadingMoreMailboxes?: boolean;
  onLoadMoreMailboxes: () => void | Promise<void>;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const configuredDomains = props.configuredDomains.map((domain) => domain.toLowerCase());
  const allDomains = uniqueStrings([...configuredDomains, ...props.mailboxes.map((mailbox) => domainForEmail(mailbox.email_address)).filter(Boolean)]).sort();
  if (allDomains.length === 0) {
    return <div className="p-6 text-center text-sm text-muted-foreground">{props.emptyDomainsText}</div>;
  }
  return (
    <>
      <div className="grid min-h-0 flex-1 content-start gap-4 overflow-auto">
        {allDomains.map((domain) => (
          <MailboxDomainGroup
            key={domain}
            {...props}
            domain={domain}
            mailboxes={props.mailboxes.filter((mailbox) => domainForEmail(mailbox.email_address) === domain)}
          />
        ))}
      </div>
      <CursorPager itemCount={props.mailboxes.length} pageSize={DEFAULT_CURSOR_PAGE_SIZE} hasNext={props.hasMoreMailboxes} loading={props.loadingMoreMailboxes} onNext={() => void props.onLoadMoreMailboxes()} />
    </>
  );
}

function MailboxDomainGroup(props: Parameters<typeof MailboxDomainGroups>[0] & { domain: string }) {
  const { hasMoreMailboxes, loadingMoreMailboxes, onLoadMoreMailboxes, ...listProps } = props;
  void hasMoreMailboxes;
  void loadingMoreMailboxes;
  void onLoadMoreMailboxes;
  return (
    <section className="grid gap-2">
      <div className="flex min-h-8 items-center justify-between border-b text-sm">
        <strong>{props.domain}</strong>
      </div>
      <MailboxRecordList {...listProps} emptyText={props.emptyDomainText} />
    </section>
  );
}
