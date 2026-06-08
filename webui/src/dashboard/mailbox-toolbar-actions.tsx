import { EyeInvisibleOutlined, EyeOutlined, InboxOutlined, KeyOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { MailboxProviderAction, type ToolbarActionDescriptor } from './dashboard-kit';
import { bulkMailboxActionCount, type MailboxProviderTab } from './mailbox-utils';
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
    label: '添加邮箱账号',
    icon: <PlusOutlined />,
    onClick: () => openImport(view.value),
  }),
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_RUN_OAUTH]: ({ action, view, props }) => {
    const count = bulkMailboxActionCount(view.mailboxes, action);
    return {
      id: 'run-oauth',
      label: '补 OAuth',
      icon: <KeyOutlined />,
      disabled: props.busy || !!props.oauthing || count === 0,
      onClick: () => void props.onOAuth(),
    };
  },
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_FETCH_INBOX]: ({ action, view, props }) => {
    const count = bulkMailboxActionCount(view.mailboxes, action);
    return {
      id: 'fetch-inbox',
      label: props.inboxLoading ? '收信中' : '拉取邮箱',
      icon: <InboxOutlined />,
      disabled: props.busy || props.inboxLoading || count === 0,
      onClick: () => void props.onFetchInbox(),
    };
  },
  [MailboxProviderAction.MAILBOX_PROVIDER_ACTION_SYNC_DOMAINS]: ({ view, props }) => ({
    id: 'sync-domains',
    label: props.domainSyncing ? '同步中' : '同步域名',
    icon: <ReloadOutlined />,
    disabled: props.busy || props.domainSyncing || !view.capability?.key,
    onClick: () => void props.onSyncDomains(view.capability?.key || ''),
  }),
};

export function providerToolbarActions(view: ProviderToolbarView, props: ProviderToolbarProps, openImport: (provider: MailboxProviderTab) => void) {
  const actions = (view.capability?.actions || [])
    .map((action) => toolbarActionFactories[action.action]?.({ action, view, props, openImport }))
    .filter((action): action is ToolbarActionDescriptor => !!action);
  return [...actions, secretsAction(props)];
}

function secretsAction(props: ProviderToolbarProps): ToolbarActionDescriptor {
  return {
    id: 'toggle-secrets',
    label: props.showSecrets ? '隐藏敏感信息' : '显示敏感信息',
    icon: props.showSecrets ? <EyeInvisibleOutlined /> : <EyeOutlined />,
    onClick: props.onToggleSecrets,
  };
}
