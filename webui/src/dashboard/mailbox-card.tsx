import { Mail } from 'lucide-react';
import { NavLink, useLocation } from 'react-router';
import {
  RecordActionButtons,
  RecordActions,
  RecordCard,
  RecordIdentity,
  RecordTop,
  StatusBadge
} from './dashboard-kit';
import { maskEmail } from './email-utils';
import { mailboxDetailPath } from './mailbox-route-paths';
import { authStatus } from './mailbox-auth-status';
import { mailboxRowActions } from './mailbox-card-actions';
import { MailboxErrorMeta, MailboxOperationMeta } from './mailbox-card-meta';
import { persistentMailboxPanelSearch } from './mailbox-panel-query';
import type { MailboxPanelMode } from './mailbox-provider-types';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';

export function MailboxCard({ mailbox, mode, selected, busy, showSecrets, oauthing, showStatus, providerCapability, currentOperation, onOAuth, onDelete }: {
  mailbox: Mailbox;
  mode: MailboxPanelMode;
  selected: boolean;
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  showStatus: boolean;
  providerCapability?: MailboxProviderCapability;
  currentOperation?: MailboxOperation;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const { search } = useLocation();
  const displayEmail = showSecrets ? mailbox.email_address : maskEmail(mailbox.email_address);
  const rowActions = mailboxRowActions({ mailbox, mode, busy, oauthing, providerCapability, currentOperation, onOAuth, onDelete });
  const detailPath = `${mailboxDetailPath(mailbox.email_address, mode === 'accounts' ? 'overview' : 'inbox')}${persistentMailboxPanelSearch(search)}`;

  return (
    <RecordCard selected={selected}>
      <NavLink to={detailPath} className="recordMain rounded-md text-inherit no-underline outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50">
        <RecordTop>
          <RecordIdentity
            icon={<Mail className="size-4" />}
            title={<span title={displayEmail}>{displayEmail}</span>}
          />
          {showStatus && <StatusBadge status={authStatus(mailbox)} />}
        </RecordTop>
        <MailboxErrorMeta error={mailbox.last_error} />
        <MailboxOperationMeta operation={currentOperation} />
      </NavLink>
      {rowActions.length > 0 && (
        <RecordActions className="rowActions">
          <div className="rowActionsMain">
            <RecordActionButtons actions={rowActions} />
          </div>
        </RecordActions>
      )}
    </RecordCard>
  );
}
