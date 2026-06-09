import { Card, EmptyBlock, Alert, AlertDescription, compactToast, formatUnix, useQuery } from './dashboard-kit';
import { formatEmailList } from './email-utils';
import { MailboxEmailViewer } from './mailbox-email-viewer';
import { fetchInboxMessageDetail, mailboxInboxMessageQueryKey } from './mailbox-inbox-query';
import { MessageSignalBadges } from './mailbox-signal-badges';
import type { InboxMessage } from './types';

export function MailboxInboxDetail({ message }: {
  message?: InboxMessage | null;
}) {
  const detailQuery = useQuery({
    queryKey: mailboxInboxMessageQueryKey(message?.mailbox_email || '', message?.id || '', message?.provider_key || ''),
    queryFn: () => message ? fetchInboxMessageDetail(message) : Promise.resolve(null),
    enabled: !!message?.mailbox_email && !!message.id
  });
  if (!message) {
    return <EmptyBlock text="选择邮件" />;
  }
  const detail = detailQuery.data;
  const detailMessage = detail?.message || message;
  const subject = detailMessage.subject || '-';
  return (
    <Card className="mailboxInboxDetail">
      <div className="mailboxInboxDetailHeader">
        <div className="min-w-0">
          <h4 className="truncate text-base font-semibold" title={subject}>{subject}</h4>
          <p className="text-xs text-muted-foreground">{formatUnix(message.received_at_unix)}</p>
        </div>
        <MessageSignalBadges message={detailMessage} className="justify-end" />
      </div>
      {detailQuery.error && (
        <Alert variant="destructive">
          <AlertDescription>{compactToast(detailQuery.error instanceof Error ? detailQuery.error.message : String(detailQuery.error))}</AlertDescription>
        </Alert>
      )}
      <dl className="mailboxInboxMetadata">
        <MetadataRow label="发件人" value={detailMessage.from_address || '-'} />
        <MetadataRow label="收件人" value={formatEmailList(detailMessage.recipients)} />
      </dl>
      <div className="mailboxInboxBody">
        {detailQuery.isFetching && <div className="mb-1 text-xs font-semibold text-muted-foreground">读取中</div>}
        <MailboxEmailViewer htmlBody={detail?.html_body || ''} textBody={detail?.body_text || detailMessage.body_preview || ''} />
      </div>
    </Card>
  );
}

function MetadataRow({ label, value, title }: { label: string; value: string; title?: string }) {
  return <><dt>{label}</dt><dd className="truncate" title={title || value}>{value}</dd></>;
}
