import { Badge, Card, EmptyBlock, formatUnix, maskPreview } from './dashboard-kit';
import { formatEmailList, maskEmail } from './email-utils';
import { messageHasVerificationSignal, messageSignals, signalHasSecretRef, signalKindName, signalLabel } from './mailbox-signal-utils';
import type { InboxMessage } from './types';

export function MailboxInboxDetail({ message, showSecrets }: {
  message?: InboxMessage | null;
  showSecrets: boolean;
}) {
  if (!message) {
    return <EmptyBlock text="选择一封邮件查看详情。" />;
  }
  const subject = showSecrets ? (message.subject || '-') : maskPreview(message.subject || '-');
  const body = showSecrets ? (message.body_preview || '-') : maskPreview(message.body_preview || '-');
  return (
    <Card className="mailboxInboxDetail">
      <div className="mailboxInboxDetailHeader">
        <div className="min-w-0">
          <h4 className="truncate text-base font-semibold" title={subject}>{subject}</h4>
          <p className="text-xs text-muted-foreground">{formatUnix(message.received_at_unix)}</p>
        </div>
        <MessageSignalBadges message={message} />
      </div>
      <dl className="mailboxInboxMetadata">
        <MetadataRow label="发件人" value={showSecrets ? (message.from_address || '-') : maskEmail(message.from_address)} />
        <MetadataRow label="收件人" value={formatEmailList(message.recipients, showSecrets)} title={formatEmailList(message.recipients, true)} />
        <MetadataRow label="来源邮箱" value={showSecrets ? (message.source_mailbox_email || message.mailbox_email || '-') : maskEmail(message.source_mailbox_email || message.mailbox_email)} />
        <MetadataRow label="Provider" value={message.provider_key || '-'} />
        <MetadataRow label="Message ID" value={message.id || '-'} />
        <MetadataRow label="大小" value={formatBytes(message.raw_size)} />
        <MetadataRow label="正文引用" value={artifactLabel(message.body_artifact_ref)} />
        <MetadataRow label="HTML 引用" value={artifactLabel(message.html_artifact_ref)} />
      </dl>
      <div className="mailboxInboxBody">
        <div className="mb-1 text-xs font-semibold text-muted-foreground">正文预览</div>
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
