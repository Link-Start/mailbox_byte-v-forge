import { Eye, EyeOff, Inbox, KeyRound, Plus, RefreshCcw } from 'lucide-react';
import { MailboxProviderAction, type ToolbarActionDescriptor } from './dashboard-kit';
import { bulkMailboxActionCount } from './mailbox-provider-capabilities';
import { mailboxAllProviderTab, type MailboxProviderTab } from './mailbox-provider-config';
import type { Mailbox, MailboxProviderActionCapability, MailboxProviderCapability } from './types';

type ProviderToolbarView = {
  value: MailboxProviderTab;
  capability?: MailboxProviderCapability;
  mailboxes: Mailbox[];
};

type ProviderToolbarProps = {
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  inboxLoading: boolean;
  domainSyncing: boolean;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onFetchInbox: () => Promise<void>;
  onSyncDomains: (providerKey: string) => Promise<void>;
  onToggleSecrets: () => void;
};

type ToolbarActionFactory = (ctx: {
  action: MailboxProviderActionCapability;
  view: ProviderToolbarView;
  props: ProviderToolbarProps;
  openImport: (provider: MailboxProviderTab) => void;
}) => ToolbarActionDescriptor;

const toolbarActionFactories: Partial<Record<MailboxProviderAction, ToolbarActionFactory>> = {
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_IMPORT_MAILBOX]: ({ view, openImport }) => ({
    id: 'import-mailbox',
    label: '添加邮箱',
    icon: <Plus className="size-4" />,
    onClick: () => openImport(view.value),
  }),
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_RUN_OAUTH]: ({ action, view, props }) => {
    const count = bulkMailboxActionCount(view.mailboxes, action);
    return {
      id: 'run-oauth',
      label: '补 OAuth',
      icon: <KeyRound className="size-4" />,
      disabled: props.busy || !!props.oauthing || count === 0,
      onClick: () => void props.onOAuth(),
    };
  },
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_FETCH_INBOX]: ({ action, view, props }) => {
    const count = bulkMailboxActionCount(view.mailboxes, action);
    return {
      id: 'fetch-inbox',
      label: props.inboxLoading ? '收信中' : '收信',
      icon: <Inbox className="size-4" />,
      disabled: props.busy || props.inboxLoading || count === 0,
      onClick: () => void props.onFetchInbox(),
    };
  },
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_SYNC_DOMAINS]: ({ view, props }) => ({
    id: 'sync-domains',
    label: props.domainSyncing ? '同步中' : '同步域名',
    icon: <RefreshCcw className="size-4" />,
    disabled: props.busy || props.domainSyncing || !view.capability?.key,
    onClick: () => void props.onSyncDomains(view.capability?.key || ''),
  }),
};

export function providerToolbarActions(view: ProviderToolbarView, props: ProviderToolbarProps, openImport: (provider: MailboxProviderTab) => void) {
  if (view.value === mailboxAllProviderTab) return aggregateToolbarActions(view, props);
  const actions = (view.capability?.actions || [])
    .map((action) => toolbarActionFactories[action.action]?.({ action, view, props, openImport }))
    .filter((action): action is ToolbarActionDescriptor => !!action);
  return [...actions, secretsAction(props)];
}

function aggregateToolbarActions(view: ProviderToolbarView, props: ProviderToolbarProps): ToolbarActionDescriptor[] {
  return [{
    id: 'fetch-all-inboxes',
    label: props.inboxLoading ? '收信中' : '收信',
    icon: <Inbox className="size-4" />,
    disabled: props.busy || props.inboxLoading || view.mailboxes.length === 0,
    onClick: () => void props.onFetchInbox(),
  }, secretsAction(props)];
}

function secretsAction(props: ProviderToolbarProps): ToolbarActionDescriptor {
  return {
    id: 'toggle-secrets',
    label: props.showSecrets ? '隐藏明文' : '显示明文',
    icon: props.showSecrets ? <EyeOff className="size-4" /> : <Eye className="size-4" />,
    onClick: props.onToggleSecrets,
  };
}
