import { Mail } from 'lucide-react';
import { NavLink, useLocation } from 'react-router';
import { Card, RecordActionButtons, StatusBadge } from './dashboard-kit';
import { mailboxDetailPath } from './mailbox-route-paths';
import { authStatus } from './mailbox-auth-status';
import { mailboxRowActions } from './mailbox-card-actions';
import { MailboxErrorMeta, MailboxOperationMeta } from './mailbox-card-meta';
import { persistentMailboxPanelSearch } from './mailbox-panel-query';
import { capabilityForProvider, providerDisplayName } from './mailbox-provider-capabilities';
import type { MailboxPanelMode } from './mailbox-provider-types';
import type { Mailbox, MailboxOperation, MailboxProviderCapability } from './types';

export function MailboxCard({ mailbox, mode, selected, busy, oauthing, showStatus, providerCapability, providerCapabilities, currentOperation, onOAuth, onDelete }: {
  mailbox: Mailbox;
  mode: MailboxPanelMode;
  selected: boolean;
  busy: boolean;
  oauthing: string;
  showStatus: boolean;
  providerCapability?: MailboxProviderCapability;
  providerCapabilities?: MailboxProviderCapability[];
  currentOperation?: MailboxOperation;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const { search } = useLocation();
  const displayEmail = mailbox.email_address;
  const rowActions = mailboxRowActions({ mailbox, mode, busy, oauthing, providerCapability, currentOperation, onOAuth, onDelete });
  const detailPath = `${mailboxDetailPath(mailbox.email_address, mode === 'accounts' ? 'overview' : 'inbox')}${persistentMailboxPanelSearch(search)}`;
  const sourceLabel = mailboxSourceLabel(mailbox, providerCapability, providerCapabilities);

  return (
    <Card className={`recordCard ${selected ? 'selected' : ''}`}>
      <NavLink to={detailPath} className="recordMain rounded-md text-inherit no-underline outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50">
        <div className="recordTop">
          <div className="recordIdentity">
            <span className="recordIcon"><Mail className="size-4" /></span>
            <div className="min-w-0">
              <strong className="recordTitle" title={displayEmail}>{displayEmail}</strong>
              <small>{sourceLabel}</small>
            </div>
          </div>
          {showStatus && <StatusBadge status={authStatus(mailbox)} />}
        </div>
        <MailboxErrorMeta error={mailbox.last_error} />
        <MailboxOperationMeta operation={currentOperation} />
      </NavLink>
      {rowActions.length > 0 && <RecordActionButtons actions={rowActions} />}
    </Card>
  );
}

function mailboxSourceLabel(mailbox: Mailbox, providerCapability?: MailboxProviderCapability, providerCapabilities: MailboxProviderCapability[] = []) {
  const capability = providerCapability || capabilityForProvider(providerCapabilities, mailbox.provider_key);
  const provider = providerDisplayName(capability, mailbox.provider_key || 'Provider');
  return [provider, mailbox.domain].filter(Boolean).join(' · ');
}
