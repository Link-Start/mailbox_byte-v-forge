import { useCallback, useState } from 'react';
import { Mail } from 'lucide-react';
import { useNavigate, useParams } from 'react-router';
import { AppDrawer, MailboxProviderAction, WorkspacePanel } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { useMailboxActions } from './mailbox-actions';
import { useMailboxData } from './mailbox-data';
import { MailboxDeleteDialog } from './mailbox-delete-dialog';
import { useMailboxEmailEventCache } from './mailbox-events';
import { MailboxPageStatus } from './mailbox-page-status';
import { mailboxDetailPath, mailboxIndexPath, type MailboxDetailTab } from './mailbox-route-paths';
import { MailboxDetails, MailboxPanel } from './mailboxes';
import { canRunProviderMailboxAction, capabilityForProvider } from './mailbox-utils';

export function MailboxPage({ detailTab = 'overview' }: { detailTab?: MailboxDetailTab }) {
  const navigate = useNavigate();
  const { mailboxEmail = '' } = useParams();
  const selectedEmail = normalizeUiEmail(mailboxEmail);
  const [showSecrets, setShowSecrets] = useState(false);
  const setSelectedEmail = useCallback((value: string | ((prev: string) => string)) => {
    const next = typeof value === 'function' ? value(selectedEmail) : value;
    void navigate(next ? mailboxDetailPath(next, detailTab) : mailboxIndexPath());
  }, [detailTab, navigate, selectedEmail]);
  const data = useMailboxData(selectedEmail);
  const actions = useMailboxActions(data, showSecrets, setSelectedEmail);
  useMailboxEmailEventCache({ email: data.selected?.email_address, inboxQueryKey: actions.inboxQueryKey, enabled: !!data.selected?.email_address });

  return (
    <>
      <WorkspacePanel>
        <div className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3">
          <MailboxPageStatus total={data.mailboxes.length} runningCount={data.runningOperationByEmail.size} showSecrets={showSecrets} error={data.loadError} />
          <MailboxPanel mailboxes={data.mailboxes} domains={data.domains} providerCapabilities={data.providerCapabilities} selected={selectedEmail} busy={data.busy} showSecrets={showSecrets} oauthing={actions.oauthing} inboxLoading={actions.inboxLoading} domainSyncing={actions.domainSyncing} runningOperationByEmail={data.runningOperationByEmail} hasMoreMailboxes={data.hasMoreMailboxes} loadingMoreMailboxes={data.loadingMoreMailboxes} onLoadMoreMailboxes={data.loadMoreMailboxes} onSelect={(mailbox) => setSelectedEmail(mailbox.email_address)} onOAuth={actions.runOAuth} onFetchInbox={() => actions.fetchInbox()} onSyncDomains={actions.syncProviderDomains} onToggleSecrets={() => setShowSecrets((value) => !value)} onDelete={actions.requestDeleteMailbox} onDone={actions.done} onError={actions.toast.showError} />
        </div>
      </WorkspacePanel>
      <AppDrawer open={!!data.selected} title="邮箱详情" icon={<Mail size={16} />} size="wide" bodyClassName="p-3" onOpenChange={(open) => { if (!open) setSelectedEmail(''); }}>
        {data.selected && <MailboxDetails mailbox={data.selected} providerCapability={capabilityForProvider(data.providerCapabilities, data.selected.provider_key)} activeTab={detailTab} showSecrets={showSecrets} inboxResult={actions.inboxResult} inboxLoading={actions.inboxLoading} canFetchInbox={canRunProviderMailboxAction(data.providerCapabilities, data.selected, MailboxProviderAction.MAILBOX_PROVIDER_ACTION_FETCH_INBOX)} onTabChange={(tab) => { if (data.selected) void navigate(mailboxDetailPath(data.selected.email_address, tab)); }} onCopy={actions.toast.copyValue} onFetchInbox={actions.fetchInbox} onDelete={actions.requestDeleteMailbox} />}
      </AppDrawer>
      <MailboxDeleteDialog mailbox={actions.deleteTarget} showSecrets={showSecrets} busy={actions.deleting} onCancel={actions.cancelDeleteMailbox} onConfirm={actions.confirmDeleteMailbox} />
    </>
  );
}
