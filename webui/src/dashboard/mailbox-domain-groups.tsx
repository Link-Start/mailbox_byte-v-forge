import { DEFAULT_CURSOR_PAGE_SIZE, CursorPager, uniqueStrings } from './dashboard-kit';
import { domainForEmail } from './mailbox-email';
import { MailboxRecordList, type MailboxRecordListProps } from './mailbox-list';

type MailboxDomainGroupsProps = Omit<MailboxRecordListProps, 'emptyText' | 'onLoadMoreMailboxes'> & {
  configuredDomains: string[];
  emptyDomainsText: string;
  emptyDomainText: string;
  onLoadMoreMailboxes: () => void | Promise<void>;
};

export function MailboxDomainGroups(props: MailboxDomainGroupsProps) {
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

function MailboxDomainGroup(props: MailboxDomainGroupsProps & { domain: string }) {
  return (
    <section className="grid gap-2">
      <div className="flex min-h-8 items-center justify-between border-b text-sm">
        <strong>{props.domain}</strong>
      </div>
      <MailboxRecordList
        mailboxes={props.mailboxes}
        emptyText={props.emptyDomainText}
        providerCapability={props.providerCapability}
        showStatus={props.showStatus}
        selected={props.selected}
        busy={props.busy}
        showSecrets={props.showSecrets}
        oauthing={props.oauthing}
        runningOperationByEmail={props.runningOperationByEmail}
        onOAuth={props.onOAuth}
        onDelete={props.onDelete}
      />
    </section>
  );
}
