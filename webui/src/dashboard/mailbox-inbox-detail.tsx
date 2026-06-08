import { Badge, Card, EmptyBlock, Alert, AlertDescription, compactToast, formatUnix, maskPreview, useQuery } from './dashboard-kit';
import { formatEmailList, maskEmail } from './email-utils';
import { fetchInboxMessageDetail, mailboxInboxMessageQueryKey } from './mailbox-inbox-query';
import { messageHasVerificationSignal, messageSignals, signalHasSecretRef, signalKindName, signalLabel } from './mailbox-signal-utils';
import type { InboxMessage } from './types';

export function MailboxInboxDetail({ message, showSecrets }: {
  message?: InboxMessage | null;
  showSecrets: boolean;
}) {
  const detailQuery = useQuery({
    queryKey: mailboxInboxMessageQueryKey(message?.mailbox_email || '', message?.id || '', message?.provider_key || ''),
    queryFn: () => message ? fetchInboxMessageDetail(message) : Promise.resolve(null),
    enabled: !!message?.mailbox_email && !!message.id
  });
  if (!message) {
    return <EmptyBlock text="选择一封邮件查看详情。" />;
  }
  const detail = detailQuery.data;
  const detailMessage = detail?.message || message;
  const subject = showSecrets ? (detailMessage.subject || '-') : maskPreview(detailMessage.subject || '-');
  const bodyText = detail?.body_text || detailMessage.body_preview || '-';
  const body = showSecrets ? bodyText : maskPreview(bodyText);
  return (
    <Card className="mailboxInboxDetail">
      <div className="mailboxInboxDetailHeader">
        <div className="min-w-0">
          <h4 className="truncate text-base font-semibold" title={subject}>{subject}</h4>
          <p className="text-xs text-muted-foreground">{formatUnix(message.received_at_unix)}</p>
        </div>
        <MessageSignalBadges message={detailMessage} />
      </div>
      {detailQuery.error && (
        <Alert variant="destructive">
          <AlertDescription>{compactToast(detailQuery.error instanceof Error ? detailQuery.error.message : String(detailQuery.error))}</AlertDescription>
        </Alert>
      )}
      <dl className="mailboxInboxMetadata">
        <MetadataRow label="发件人" value={showSecrets ? (detailMessage.from_address || '-') : maskEmail(detailMessage.from_address)} />
        <MetadataRow label="收件人" value={formatEmailList(detailMessage.recipients, showSecrets)} title={formatEmailList(detailMessage.recipients, true)} />
        <MetadataRow label="来源邮箱" value={showSecrets ? (detailMessage.source_mailbox_email || detailMessage.mailbox_email || '-') : maskEmail(detailMessage.source_mailbox_email || detailMessage.mailbox_email)} />
        <MetadataRow label="Provider" value={detailMessage.provider_key || '-'} />
        <MetadataRow label="Message ID" value={detailMessage.id || '-'} />
        <MetadataRow label="大小" value={formatBytes(detailMessage.raw_size)} />
        <MetadataRow label="正文引用" value={artifactLabel(detailMessage.body_artifact_ref)} />
        <MetadataRow label="HTML 引用" value={detail?.html_body ? `${formatBytes(detail.html_body.length)} · 未渲染` : artifactLabel(detailMessage.html_artifact_ref)} />
      </dl>
      <div className="mailboxInboxBody">
        <div className="mb-1 text-xs font-semibold text-muted-foreground">{detailQuery.isFetching ? '正在读取正文' : '正文'}</div>
        <pre>{body}</pre>
      </div>
    </Card>
  );
}

function MetadataRow({ label, value, title }: { label: string; value: string; title?: string }) {
  return <><dt>{label}</dt><dd className="truncate" title={title || value}>{value}</dd></>;
}

function MessageSignalBadges({ message }: { message: InboxMessage }) {
  const signals = messageSignals(message);
  if (signals.length === 0 && messageHasVerificationSignal(message)) {
    return <Badge variant="outline" className="border-emerald-200 bg-emerald-50 text-emerald-700">验证码 已检测</Badge>;
  }
  if (signals.length === 0) return null;
  return (
    <span className="flex shrink-0 flex-wrap items-center justify-end gap-1">
      {signals.map((signal, index) => {
        const kind = signalKindName(signal.kind);
        const captured = kind === 'otp' && signalHasSecretRef(signal, 'otp');
        return <Badge variant="secondary" key={`${kind}-${signal.label || index}`}>{signalLabel(signal)}{captured ? ' 已保存' : ''}</Badge>;
      })}
    </span>
  );
}

function artifactLabel(ref: InboxMessage['body_artifact_ref']) {
  if (!ref?.artifact_id) return '-';
  return `${ref.purpose || 'artifact'} · ${formatBytes(ref.size_bytes)}`;
}

function formatBytes(value: number) {
  if (!value || value <= 0) return '-';
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}
