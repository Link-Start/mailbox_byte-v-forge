import { Interweave } from 'interweave';
import { Card, EmptyBlock, Alert, AlertDescription, compactToast, formatUnix, useQuery } from './dashboard-kit';
import { formatEmailList } from './email-utils';
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
  const body = emailBodyContent(detail?.body_text || detailMessage.body_preview || '', detail?.html_body || '');
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
        <Interweave className="mailboxEmailViewer" content={body} />
      </div>
    </Card>
  );
}

function MetadataRow({ label, value, title }: { label: string; value: string; title?: string }) {
  return <><dt>{label}</dt><dd className="truncate" title={title || value}>{value}</dd></>;
}

function emailBodyContent(textBody: string, htmlBody: string) {
  const html = String(htmlBody || '').trim();
  if (html) return html;
  return textAsHTML(cleanTextBody(textBody) || '-');
}

function cleanTextBody(value: string) {
  return String(value || '')
    .replace(/\r\n/g, '\n')
    .replace(/([^\s<][^<\n]{0,120})<https?:\/\/[^>\s]+>/g, '$1')
    .replace(/<https?:\/\/[^>\s]+>/g, '')
    .replace(/\n{3,}/g, '\n\n')
    .trim();
}

function textAsHTML(value: string) {
  return escapeHTML(value).replace(/\n/g, '<br>');
}

function escapeHTML(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}
