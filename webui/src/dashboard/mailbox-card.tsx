import { AlertCircle, KeyRound, Mail, Trash2 } from 'lucide-react';
import {
  RecordActionButtons,
  RecordActions,
  RecordCard,
  RecordIdentity,
  RecordMain,
  RecordMeta,
  RecordTop,
  StatusBadge,
  MailboxProviderAction,
  type RowActionDescriptor
} from './dashboard-kit';
import { maskEmail } from './email-utils';
import { authStatus, canRunMailboxAction, providerAction } from './mailbox-utils';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';

export function MailboxCard({ mailbox, selected, busy, showSecrets, oauthing, showStatus, providerCapability, currentOperation, onSelect, onOAuth, onDelete }: {
  mailbox: Mailbox;
  selected: boolean;
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  showStatus: boolean;
  providerCapability?: MailboxProviderCapability;
  currentOperation?: MailboxOperation;
  onSelect: (mailbox: Mailbox) => void;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const displayEmail = showSecrets ? mailbox.email_address : maskEmail(mailbox.email_address);
  const rowActions = mailboxRowActions({ mailbox, busy, oauthing, providerCapability, currentOperation, onOAuth, onDelete });

  return (
    <RecordCard selected={selected} onClick={() => onSelect(mailbox)}>
      <RecordMain>
        <RecordTop>
          <RecordIdentity
            icon={<Mail className="size-4" />}
            title={<span title={displayEmail}>{displayEmail}</span>}
          />
          {showStatus && <StatusBadge status={authStatus(mailbox)} />}
        </RecordTop>
        <MailboxErrorMeta error={mailbox.last_error} />
        <MailboxOperationMeta operation={currentOperation} />
      </RecordMain>
      <RecordActions className="rowActions">
        <div className="rowActionsMain">
          <RecordActionButtons actions={rowActions} />
        </div>
      </RecordActions>
    </RecordCard>
  );
}

function mailboxRowActions({ mailbox, busy, oauthing, providerCapability, currentOperation, onOAuth, onDelete }: {
  mailbox: Mailbox;
  busy: boolean;
  oauthing: string;
  providerCapability?: MailboxProviderCapability;
  currentOperation?: MailboxOperation;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
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

function MailboxErrorMeta({ error }: { error?: string }) {
  if (!error) return null;
  return (
    <RecordMeta className="grid-cols-1">
      <span className="flex min-w-0 items-center gap-1 truncate text-xs text-destructive" title={error}>
        <AlertCircle className="size-3.5 shrink-0" />
        {error}
      </span>
    </RecordMeta>
  );
}

function MailboxOperationMeta({ operation }: { operation?: MailboxOperation }) {
  if (!operation) return null;
  return (
    <RecordMeta className="grid-cols-1">
      <span className="truncate text-xs text-muted-foreground" title={operation.operation_id}>
        运行中 · {operation.last_step || operation.action || operation.status}
      </span>
    </RecordMeta>
  );
}
