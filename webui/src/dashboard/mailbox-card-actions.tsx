import { KeyRound, Trash2 } from 'lucide-react';
import { MailboxProviderAction, type RowActionDescriptor } from './dashboard-kit';
import { canRunMailboxAction, providerAction } from './mailbox-provider-capabilities';
import type { MailboxPanelMode } from './mailbox-provider-types';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';

export function mailboxRowActions({ mailbox, mode, busy, oauthing, providerCapability, currentOperation, onOAuth, onDelete }: {
  mailbox: Mailbox;
  mode: MailboxPanelMode;
  busy: boolean;
  oauthing: string;
  providerCapability?: MailboxProviderCapability;
  currentOperation?: MailboxOperation;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  if (mode !== 'accounts') return [];
  const actions: RowActionDescriptor[] = [{
    id: 'delete-mailbox',
    label: '删除邮箱',
    icon: <Trash2 className="size-4" />,
    disabled: busy || !!oauthing,
    kind: 'danger',
    onClick: () => void onDelete(mailbox),
  }];
  const canOAuth = canRunMailboxAction(mailbox, providerAction(providerCapability, MailboxProviderAction.MAILBOX_PROVIDER_ACTION_RUN_OAUTH));
  if (!currentOperation && canOAuth) {
    actions.unshift({
      id: 'run-oauth',
      label: oauthing === mailbox.email_address || oauthing === '*' ? 'OAuth 提交中' : '补 OAuth',
      icon: <KeyRound className="size-4" />,
      disabled: busy || !!oauthing,
      onClick: () => void onOAuth(mailbox.email_address),
    });
  }
  return actions;
}
