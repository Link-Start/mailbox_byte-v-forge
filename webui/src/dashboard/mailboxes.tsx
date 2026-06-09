import { Inbox, Search, UsersRound, X } from 'lucide-react';
import {
  Badge,
  Button,
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
  EmptyBlock,
  Input,
  PanelTabs,
  ToolbarActionButtons
} from './dashboard-kit';
import { MailboxImportSheet } from './mailbox-import';
import type { MailboxPanelMode, MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerToolbarActions } from './mailbox-toolbar-actions';
import { capabilityForProvider, providerShowsCredentialState } from './mailbox-provider-capabilities';
import { mailboxAllProviderTab, type MailboxProviderTab } from './mailbox-provider-config';
import { useMailboxPanelState } from './mailbox-panel-state';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';
import { MailboxRecordList } from './mailbox-list';
export { MailboxDetails } from './mailbox-details';

export function MailboxPanel(props: MailboxPanelProps) {
  const state = useMailboxPanelState(props.mailboxes, props.providerCapabilities);
  if (state.providerViews.length === 0) return <EmptyBlock text={props.busy ? '加载中' : '暂无 Provider'} />;
  if (props.mode === 'inbox') return <InboxMailboxPanel props={props} state={state} />;
  return <AccountsMailboxPanel props={props} state={state} />;
}

type MailboxPanelState = ReturnType<typeof useMailboxPanelState>;

type MailboxPanelProps = {
  mailboxes: Mailbox[];
  mode: MailboxPanelMode;
  providerCapabilities: MailboxProviderCapability[];
  selected?: string;
  busy: boolean;
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
  onDelete: (mailbox: Mailbox) => void;
  onDone: (message: string) => void;
  onError: (message: string) => void;
};

function InboxMailboxPanel({ props, state }: { props: MailboxPanelProps; state: MailboxPanelState }) {
  const actions = providerToolbarActions({ value: mailboxAllProviderTab, mailboxes: state.filteredMailboxes }, props, () => undefined);
  return (
    <div className="mailboxListPanel">
      <MailboxPanelSearch props={props} state={state} />
      <Card className="mailboxPanelCard">
        <CardHeader className="mailboxPanelHeader">
          <CardTitle className="panelTitle"><Inbox className="size-4" />收件箱</CardTitle>
          <CardAction><ToolbarActionButtons actions={actions} /></CardAction>
        </CardHeader>
        <CardContent className="mailboxPanelContent">
          <MailboxRecordList
            {...providerPanelProps(props, state.searchQuery)}
            mailboxes={state.filteredMailboxes}
            showStatus={false}
            emptyText={state.searchQuery ? '无匹配邮箱' : '暂无邮箱'}
          />
        </CardContent>
      </Card>
    </div>
  );
}

function AccountsMailboxPanel({ props, state }: { props: MailboxPanelProps; state: MailboxPanelState }) {
  const panelProps = providerPanelProps(props, state.searchQuery);
  return (
    <div className="mailboxListPanel">
      <MailboxPanelSearch props={props} state={state} />
      <PanelTabs
        value={state.activeProvider}
        onValueChange={(value) => state.updateQuery({ provider: value as MailboxProviderTab })}
        tabsClassName="min-h-0 flex-1 overflow-hidden"
        tabsListVariant="line"
        tabsListClassName="h-8"
        tabs={state.providerViews.map(({ value, label, capability, mailboxes }) => ({
          value,
          label,
          triggerClassName: 'gap-1.5 px-2',
          contentClassName: 'flex flex-col overflow-hidden',
          content: (
            <Card className="mailboxPanelCard">
              <CardHeader className="mailboxPanelHeader mailboxPanelHeaderCompact">
                <CardTitle className="panelTitle">{label}</CardTitle>
                <CardAction>
                  <ToolbarActionButtons
                    actions={providerToolbarActions({ value, capability, mailboxes }, props, (provider) => state.updateQuery({ importProvider: provider }))}
                  />
                </CardAction>
              </CardHeader>
              <CardContent className="mailboxPanelContent">
                <MailboxRecordList
                  {...panelProps}
                  mailboxes={mailboxes}
                  providerCapability={capability}
                  showStatus={providerShowsCredentialState(capability)}
                  emptyText={state.searchQuery ? '无匹配邮箱' : '暂无邮箱'}
                />
              </CardContent>
            </Card>
          )
        }))}
      />
      <MailboxImportSheet
        open={!!state.importProvider}
        provider={state.importProvider || state.activeProvider}
        capability={capabilityForProvider(props.providerCapabilities, state.importProvider || state.activeProvider)}
        busy={props.busy}
        onOpenChange={(open) => !open && state.updateQuery({ importProvider: '' })}
        onDone={props.onDone}
        onError={props.onError}
      />
    </div>
  );
}

function MailboxPanelSearch({ props, state }: { props: MailboxPanelProps; state: MailboxPanelState }) {
  return (
    <div className="mailboxSearchBar">
      <UsersRound className="hidden size-4 text-muted-foreground sm:block" />
      <div className="relative min-w-[220px] flex-1">
        <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          className="pl-9 pr-9"
          value={state.query}
          placeholder="搜索邮箱、域名、Provider"
          onChange={(event) => state.updateQuery({ q: event.target.value })}
        />
        {state.query && (
          <Button
            className="absolute right-1 top-1/2 size-7 -translate-y-1/2"
            variant="ghost"
            size="icon"
            aria-label="清空搜索"
            onClick={() => state.updateQuery({ q: '' })}
          >
            <X className="size-4" />
          </Button>
        )}
      </div>
      {state.searchQuery && <Badge variant="secondary">{state.filteredMailboxes.length}/{props.mailboxes.length}</Badge>}
    </div>
  );
}

function providerPanelProps(props: MailboxPanelProps, searchQuery: string): Omit<MailboxProviderPanelProps, 'mailboxes' | 'capability'> {
  return {
    mode: props.mode,
    providerCapabilities: props.providerCapabilities,
    selected: props.selected,
    busy: props.busy,
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
    onDelete: props.onDelete,
    onDone: props.onDone,
    onError: props.onError
  };
}
