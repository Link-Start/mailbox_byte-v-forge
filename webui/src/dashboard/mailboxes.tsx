import { useEffect, useMemo, useState } from 'react';
import { Search, X } from 'lucide-react';
import {
  Badge,
  Button,
  EmptyBlock,
  Input,
  PanelTabs,
  ToolbarActionButtons
} from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { MailboxImportSheet } from './mailbox-import';
import { mailboxProviderViews } from './mailbox-provider-registry';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerToolbarActions } from './mailbox-toolbar-actions';
import { capabilityForProvider } from './mailbox-provider-capabilities';
import type { MailboxProviderTab } from './mailbox-provider-config';
import type { Mailbox, MailboxDomain, MailboxOperation, MailboxProviderCapability } from './types';
export { MailboxDetails } from './mailbox-details';

export function MailboxPanel(props: MailboxPanelProps) {
  const [activeProvider, setActiveProvider] = useState<MailboxProviderTab>('');
  const [importProvider, setImportProvider] = useState<MailboxProviderTab>();
  const [query, setQuery] = useState('');
  const searchQuery = query.trim();
  const filteredMailboxes = useMemo(() => filterMailboxes(props.mailboxes, searchQuery), [props.mailboxes, searchQuery]);
  const panelProps = providerPanelProps(props, searchQuery);
  const providerViews = useMemo(() => mailboxProviderViews(props.providerCapabilities, filteredMailboxes), [props.providerCapabilities, filteredMailboxes]);
  useEffect(() => {
    if (providerViews.length > 0 && !providerViews.some((view) => view.value === activeProvider)) setActiveProvider(providerViews[0].value);
  }, [activeProvider, providerViews]);
  if (providerViews.length === 0) return <EmptyBlock text={props.busy ? '加载中' : '暂无 Provider'} />;
  return (
    <div className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <div className="relative min-w-[240px] flex-1">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input className="pl-9 pr-9" value={query} placeholder="搜索" onChange={(event) => setQuery(event.target.value)} />
          {query && <Button className="absolute right-1 top-1/2 size-7 -translate-y-1/2" variant="ghost" size="icon" aria-label="清空搜索" onClick={() => setQuery('')}><X className="size-4" /></Button>}
        </div>
        {searchQuery && <Badge variant="secondary">{filteredMailboxes.length}/{props.mailboxes.length}</Badge>}
      </div>
      <PanelTabs
        value={activeProvider}
        onValueChange={(value) => setActiveProvider(value as MailboxProviderTab)}
        tabsClassName="min-h-0 flex-1 overflow-hidden"
        tabsListVariant="line"
        tabsListClassName="h-8"
        tabs={providerViews.map(({ value, label, capability, mailboxes, Component }) => ({
          value,
          label,
          triggerClassName: 'gap-1.5 px-2',
          contentClassName: 'flex flex-col overflow-hidden',
          content: (
            <Component
              {...panelProps}
              mailboxes={mailboxes}
              capability={capability}
              actions={<ToolbarActionButtons actions={providerToolbarActions({ value, capability, mailboxes }, props, setImportProvider)} />}
            />
          )
        }))}
      />
      <MailboxImportSheet open={!!importProvider} provider={importProvider || activeProvider} capability={capabilityForProvider(props.providerCapabilities, importProvider || activeProvider)} busy={props.busy} onOpenChange={(open) => !open && setImportProvider(undefined)} onDone={props.onDone} onError={props.onError} />
    </div>
  );
}

type MailboxPanelProps = {
  mailboxes: Mailbox[];
  domains: MailboxDomain[];
  providerCapabilities: MailboxProviderCapability[];
  selected?: string;
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  inboxLoading: boolean;
  domainSyncing: boolean;
  runningOperationByEmail: Map<string, MailboxOperation>;
  hasMoreMailboxes?: boolean;
  loadingMoreMailboxes?: boolean;
  onLoadMoreMailboxes: () => void | Promise<void>;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onFetchInbox: () => Promise<void>;
  onSyncDomains: (providerKey: string) => Promise<void>;
  onToggleSecrets: () => void;
  onDelete: (mailbox: Mailbox) => void;
  onDone: (message: string) => void;
  onError: (message: string) => void;
};

function providerPanelProps(props: MailboxPanelProps, searchQuery: string): Omit<MailboxProviderPanelProps, 'mailboxes' | 'capability'> {
  return {
    domains: props.domains,
    selected: props.selected,
    busy: props.busy,
    showSecrets: props.showSecrets,
    oauthing: props.oauthing,
    inboxLoading: props.inboxLoading,
    domainSyncing: props.domainSyncing,
    runningOperationByEmail: props.runningOperationByEmail,
    searchQuery,
    hasMoreMailboxes: props.hasMoreMailboxes,
    loadingMoreMailboxes: props.loadingMoreMailboxes,
    onLoadMoreMailboxes: props.onLoadMoreMailboxes,
    onOAuth: props.onOAuth,
    onFetchInbox: props.onFetchInbox,
    onSyncDomains: props.onSyncDomains,
    onToggleSecrets: props.onToggleSecrets,
    onDelete: props.onDelete,
    onDone: props.onDone,
    onError: props.onError
  };
}

function filterMailboxes(mailboxes: Mailbox[], query: string) {
  const needle = query.trim().toLowerCase();
  if (!needle) return mailboxes;
  return mailboxes.filter((mailbox) => [
    mailbox.email_address,
    mailbox.domain,
    mailbox.provider_key
  ].some((value) => normalizeUiEmail(value || '').includes(needle)));
}
