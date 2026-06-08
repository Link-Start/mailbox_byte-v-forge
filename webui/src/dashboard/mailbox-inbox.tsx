import { useEffect, useMemo, useState } from 'react';
import { Inbox } from 'lucide-react';
import {
  Alert,
  AlertDescription,
  Badge,
  Button,
  EmptyBlock,
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
  compactToast,
  formatUnix,
  maskPreview
} from './dashboard-kit';
import { formatEmailList, maskEmail } from './email-utils';
import { MailboxInboxDetail } from './mailbox-inbox-detail';
import { messageHasVerificationSignal, messageSignals, signalHasSecretRef, signalKindName, signalLabel } from './mailbox-signal-utils';
import type { InboxMessage, InboxResult, Mailbox } from './types';

export function MailboxInboxSection({ mailbox, result, showSecrets, loading, canFetch, onFetch }: {
  mailbox: Mailbox;
  result?: InboxResult | null;
  showSecrets: boolean;
  loading: boolean;
  canFetch: boolean;
  onFetch: (emailAddress?: string) => Promise<void>;
}) {
  const messages = useMemo(() => result?.messages || [], [result?.messages]);
  const [selectedKey, setSelectedKey] = useState('');
  const selectedMessage = messages.find((message, index) => inboxMessageKey(message, index) === selectedKey) || null;

  useEffect(() => {
    if (messages.length === 0) {
      setSelectedKey('');
      return;
    }
    if (!messages.some((message, index) => inboxMessageKey(message, index) === selectedKey)) {
      setSelectedKey(inboxMessageKey(messages[0], 0));
    }
  }, [messages, selectedKey]);

  return (
    <section className="grid gap-3">
      <div className="flex items-center justify-between gap-2">
        <h3 className="text-sm font-semibold">收件箱</h3>
        {canFetch && (
          <Button variant="outline" size="sm" disabled={loading} onClick={() => onFetch(mailbox.email_address)}>
            <Inbox />{loading ? '刷新中' : '刷新'}
          </Button>
        )}
      </div>
      {result?.error_message && (
        <Alert variant="destructive">
          <AlertDescription>{compactToast(result.error_message)}</AlertDescription>
        </Alert>
      )}
      <div className="mailboxInboxLayout">
        <div className="grid gap-2">
          {messages.map((message, index) => {
            const key = inboxMessageKey(message, index);
            return <InboxMessageRow message={message} selected={key === selectedKey} showSecrets={showSecrets} key={key} onSelect={() => setSelectedKey(key)} />;
          })}
          {!result && <EmptyBlock text={loading ? '正在读取收件箱。' : '暂无邮件。'} />}
          {result && !result.error_message && messages.length === 0 && <EmptyBlock text="当前邮箱没有新邮件。" />}
        </div>
        {messages.length > 0 && <MailboxInboxDetail message={selectedMessage} showSecrets={showSecrets} />}
      </div>
    </section>
  );
}

function InboxMessageRow({ message, selected, showSecrets, onSelect }: {
  message: InboxMessage;
  selected: boolean;
  showSecrets: boolean;
  onSelect: () => void;
}) {
  return (
    <Item variant="outline" className={`inboxMessageRow items-start ${selected ? 'selected' : ''}`} role="button" tabIndex={0} onClick={onSelect} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') onSelect(); }}>
      <ItemContent className="min-w-0">
        <ItemTitle className="w-full justify-between gap-2">
          <span className="truncate" title={message.subject}>{message.subject || '-'}</span>
          <span className="shrink-0 text-xs font-normal text-muted-foreground">{formatUnix(message.received_at_unix)}</span>
        </ItemTitle>
        <ItemDescription className="flex items-center justify-between gap-2">
          <span className="truncate">发件人 {showSecrets ? (message.from_address || '-') : maskEmail(message.from_address)}</span>
          <MessageSignalStrip message={message} />
        </ItemDescription>
        <ItemDescription className="line-clamp-1" title={formatEmailList(message.recipients, true)}>
          收件人 {formatEmailList(message.recipients, showSecrets)}
        </ItemDescription>
        <ItemDescription className="line-clamp-3">{showSecrets ? (message.body_preview || '-') : maskPreview(message.body_preview || '-')}</ItemDescription>
      </ItemContent>
    </Item>
  );
}

function MessageSignalStrip({ message }: { message: InboxMessage }) {
  const signals = messageSignals(message);
  const fallbackSignal = messageHasVerificationSignal(message);
  if (signals.length === 0 && fallbackSignal) {
    return <Badge variant="outline" className="border-emerald-200 bg-emerald-50 text-emerald-700">验证码 已捕获</Badge>;
  }
  if (signals.length === 0) return null;
  return (
    <span className="flex shrink-0 items-center gap-1">
      {signals.map((signal, index) => {
        const kind = signalKindName(signal.kind);
        const captured = kind === 'otp' && signalHasSecretRef(signal, 'otp');
        return <Badge variant="secondary" key={`${kind}-${signal.label || index}`}>{signalLabel(signal)}{captured ? ' 已捕获' : ''}</Badge>;
      })}
    </span>
  );
}

function inboxMessageKey(message: InboxMessage, index: number) {
  return [message.provider_key || '', message.mailbox_email || '', message.id || index, message.received_at_unix || 0].join(':');
}
