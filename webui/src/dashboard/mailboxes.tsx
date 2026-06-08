import { useEffect, useMemo, useState } from 'react';
import {
  EmptyBlock,
  PanelTabs,
  ToolbarActionButtons
} from './dashboard-kit';
import { MailboxImportSheet } from './mailbox-import';
import { mailboxProviderViews } from './mailbox-provider-registry';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerToolbarActions } from './mailbox-toolbar-actions';
import { capabilityForProvider, type MailboxProviderTab } from './mailbox-utils';
import type { Mailbox, MailboxDomain, MailboxOperation, MailboxProviderCapability } from './types';
export { MailboxDetails } from './mailbox-details';

export function MailboxPanel(props: MailboxPanelProps) {
  const [activeProvider, setActiveProvider] = useState<MailboxProviderTab>('');
  const [importProvider, setImportProvider] = useState<MailboxProviderTab>();
  const panelProps = providerPanelProps(props);
  const providerViews = useMemo(() => mailboxProviderViews(props.providerCapabilities, props.mailboxes), [props.providerCapabilities, props.mailboxes]);
  useEffect(() => {
    if (providerViews.length > 0 && !providerViews.some((view) => view.value === activeProvider)) setActiveProvider(providerViews[0].value);
  }, [activeProvider, providerViews]);
  if (providerViews.length === 0) return <EmptyBlock text="暂无可用邮箱 provider。" />;
  return (
    <>
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
    </>
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
  onSelect: (mailbox: Mailbox) => void;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onFetchInbox: () => Promise<void>;
  onSyncDomains: (providerKey: string) => Promise<void>;
  onToggleSecrets: () => void;
  onDelete: (mailbox: Mailbox) => void;
  onDone: (message: string) => void;
  onError: (message: string) => void;
};

function providerPanelProps(props: MailboxPanelProps): Omit<MailboxProviderPanelProps, 'mailboxes' | 'capability'> {
  return {
    domains: props.domains,
    selected: props.selected,
    busy: props.busy,
    showSecrets: props.showSecrets,
    oauthing: props.oauthing,
    inboxLoading: props.inboxLoading,
    domainSyncing: props.domainSyncing,
    runningOperationByEmail: props.runningOperationByEmail,
    hasMoreMailboxes: props.hasMoreMailboxes,
    loadingMoreMailboxes: props.loadingMoreMailboxes,
    onLoadMoreMailboxes: props.onLoadMoreMailboxes,
    onSelect: props.onSelect,
    onOAuth: props.onOAuth,
    onFetchInbox: props.onFetchInbox,
    onSyncDomains: props.onSyncDomains,
    onToggleSecrets: props.onToggleSecrets,
    onDelete: props.onDelete,
    onDone: props.onDone,
    onError: props.onError
  };
}
