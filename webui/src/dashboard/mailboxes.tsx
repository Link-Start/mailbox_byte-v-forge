import { useState } from 'react';
import type { ComponentType } from 'react';
import {
  PanelTabs,
  ToolbarActionButtons
} from './dashboard-kit';
import { CloudflareMailboxProviderPanel } from './mailbox-provider-cloudflare';
import { MailboxImportSheet } from './mailbox-import';
import { OutlookMailboxProviderPanel } from './mailbox-provider-outlook';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerToolbarActions } from './mailbox-toolbar-actions';
import { capabilityForProvider, mailboxProviderMatches, type MailboxProviderTab } from './mailbox-utils';
import type { Mailbox, MailboxDomain, MailboxOperation, MailboxProviderCapability } from './types';
export { MailboxDetails } from './mailbox-details';

const providerComponents: ProviderDefinition[] = [{
  value: 'outlook',
  fallbackLabel: 'Outlook',
  Component: OutlookMailboxProviderPanel,
}, {
  value: 'cloudflare',
  fallbackLabel: 'Cloudflare',
  Component: CloudflareMailboxProviderPanel,
}];

export function MailboxPanel(props: MailboxPanelProps) {
  const [activeProvider, setActiveProvider] = useState<MailboxProviderTab>('outlook');
  const [importProvider, setImportProvider] = useState<MailboxProviderTab>();
  const panelProps = providerPanelProps(props);
  const providerViews = providerDefinitions(props.providerCapabilities).map((definition) => ({
    ...definition,
    capability: capabilityForProvider(props.providerCapabilities, definition.value),
    mailboxes: props.mailboxes.filter((mailbox) => mailboxProviderMatches(mailbox.provider_key, definition.value)),
  }));
  return (
    <>
      <PanelTabs
        value={activeProvider}
        onValueChange={(value) => setActiveProvider(value as MailboxProviderTab)}
        tabsClassName="min-h-0 flex-1 overflow-hidden"
        tabsListVariant="line"
        tabsListClassName="h-8"
        tabs={providerViews.map(({ value, fallbackLabel, capability, mailboxes, Component }) => ({
          value,
          label: capability?.display_name || fallbackLabel,
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

type ProviderDefinition = {
  value: MailboxProviderTab;
  fallbackLabel: string;
  Component: ComponentType<MailboxProviderPanelProps>;
};

function providerDefinitions(capabilities: MailboxProviderCapability[]): ProviderDefinition[] {
  const fromCapabilities = capabilities
    .map((capability) => providerComponents.find((item) => mailboxProviderMatches(capability.key, item.value)))
    .filter((item): item is ProviderDefinition => !!item);
  return fromCapabilities.length > 0 ? fromCapabilities : providerComponents;
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
  onDelete: (mailbox: Mailbox) => Promise<void>;
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
