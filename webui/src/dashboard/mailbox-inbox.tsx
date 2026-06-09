import { useDeferredValue, useEffect, useState } from 'react';
import { Mail, RefreshCcw, Search } from 'lucide-react';
import {
  Alert,
  AlertDescription,
  Badge,
  Button,
  CursorPager,
  EmptyBlock,
  Input,
  Item,
  ItemContent,
  ItemDescription,
  ItemTitle,
  compactToast,
  formatInboxTime,
  formatUnix,
  ScrollArea
} from './dashboard-kit';
import { emailDisplayName, emailInitial } from './email-utils';
import { useMailboxEmailEventCache } from './mailbox-events';
import { MailboxInboxDetail } from './mailbox-inbox-detail';
import { mailboxInboxPageSize, mailboxInboxPagesQueryKey, useMailboxInboxPages } from './mailbox-inbox-pages';
import { MessageSignalBadges } from './mailbox-signal-badges';
import type { InboxMessage, Mailbox } from './types';

export function MailboxInboxSection({ mailbox, loading, canFetch, onFetch }: {
  mailbox: Mailbox;
  loading: boolean;
  canFetch: boolean;
  onFetch: (emailAddress?: string) => Promise<void>;
}) {
  const [searchQuery, setSearchQuery] = useState('');
  const deferredSearch = useDeferredValue(searchQuery);
  const inbox = useMailboxInboxPages(mailbox.email_address, deferredSearch);
  useMailboxEmailEventCache({
    email: mailbox.email_address,
    signalKind: 'any',
    inboxQueryKey: mailboxInboxPagesQueryKey(mailbox.email_address, deferredSearch)
  });
  const messages = inbox.messages;
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
    <section className="grid h-full min-h-0 gap-3">
      {inbox.errorMessage && (
        <Alert variant="destructive">
          <AlertDescription>{compactToast(inbox.errorMessage)}</AlertDescription>
        </Alert>
      )}
      <div className="mailboxInboxLayout">
        <div className="mailboxInboxListPane">
          <div className="mailboxInboxToolbar">
            <div className="flex min-w-0 items-center gap-2">
              <Mail className="size-4 text-muted-foreground" />
              <strong>邮件</strong>
              <Badge variant="secondary">{messages.length}</Badge>
            </div>
            <div className="relative min-w-0 flex-1">
              <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input className="h-8 pl-8" value={searchQuery} placeholder="搜索邮件" onChange={(event) => setSearchQuery(event.target.value)} />
            </div>
            {canFetch && <Button variant="ghost" size="icon" title={loading ? '刷新中' : '刷新'} aria-label={loading ? '刷新中' : '刷新'} disabled={loading} onClick={() => onFetch(mailbox.email_address)}><RefreshCcw className="size-4" /></Button>}
          </div>
          <ScrollArea className="h-full min-h-0">
            <div className="grid gap-2 pr-2">
              {messages.map((message, index) => {
                const key = inboxMessageKey(message, index);
                return <InboxMessageRow message={message} selected={key === selectedKey} key={key} onSelect={() => setSelectedKey(key)} />;
              })}
              {inbox.isLoading && <EmptyBlock text="读取中" />}
              {!inbox.isLoading && !inbox.errorMessage && messages.length === 0 && <EmptyBlock text={deferredSearch ? '无匹配邮件' : '暂无邮件'} />}
              <CursorPager itemCount={messages.length} pageSize={mailboxInboxPageSize} hasNext={inbox.hasNextPage} loading={inbox.isFetchingNextPage} onNext={inbox.loadMore} />
            </div>
          </ScrollArea>
        </div>
        {messages.length > 0 && <MailboxInboxDetail message={selectedMessage} />}
      </div>
    </section>
  );
}

function InboxMessageRow({ message, selected, onSelect }: {
  message: InboxMessage;
  selected: boolean;
  onSelect: () => void;
}) {
  const from = message.from_address || '-';
  const preview = message.body_preview || '-';
  const isRecent = Date.now() / 1000 - (message.received_at_unix || 0) < 300;
  return (
    <Item
      className={`inboxMessageRow items-start ${selected ? 'selected' : ''}`}
      role="button"
      tabIndex={0}
      aria-pressed={selected}
      onClick={onSelect}
      onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') onSelect(); }}
    >
      <span className="inboxSenderAvatar" aria-hidden="true">{emailInitial(message.from_address)}</span>
      <ItemContent className="min-w-0">
        <ItemTitle className="inboxMessageTop">
          <span className="truncate" title={from}>{emailDisplayName(message.from_address)}</span>
          <span className="shrink-0 text-xs font-normal text-muted-foreground" title={formatUnix(message.received_at_unix)}>
            {formatInboxTime(message.received_at_unix)}
          </span>
        </ItemTitle>
        <ItemDescription className="inboxMessageSubject">
          <span className="truncate" title={message.subject}>{message.subject || '-'}</span>
          {isRecent && <Badge className="inboxNewBadge" variant="secondary">新</Badge>}
        </ItemDescription>
        <ItemDescription className="inboxMessageMeta">
          <span className="line-clamp-2 leading-relaxed">{preview}</span>
          <MessageSignalBadges message={message} />
        </ItemDescription>
      </ItemContent>
    </Item>
  );
}

function inboxMessageKey(message: InboxMessage, index: number) {
  return [message.provider_key || '', message.mailbox_email || '', message.id || index, message.received_at_unix || 0].join(':');
}
