import { useCallback, useState } from 'react';
import { Outlet, useLocation, useNavigate, useOutletContext, useParams } from 'react-router';
import { MailboxProviderAction, WorkspacePanel } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { useMailboxActions } from './mailbox-actions';
import { useMailboxData } from './mailbox-data';
import { MailboxDeleteDialog } from './mailbox-delete-dialog';
import { useMailboxEmailEventCache } from './mailbox-events';
import { MailboxPageStatus } from './mailbox-page-status';
import { persistentMailboxPanelSearch } from './mailbox-panel-query';
import { MailboxReadingEmpty, MailboxReadingPane } from './mailbox-reading-pane';
import { mailboxDetailPath, mailboxInboxIndexPath, type MailboxDetailTab } from './mailbox-route-paths';
import { MailboxPanel } from './mailboxes';
import { canRunProviderMailboxAction, capabilityForProvider } from './mailbox-provider-capabilities';

type MailboxPageContext = {
  data: ReturnType<typeof useMailboxData>;
  actions: ReturnType<typeof useMailboxActions>;
  showSecrets: boolean;
  closeDetails: () => void;
  openDetailTab: (tab: MailboxDetailTab) => void;
};

export function MailboxPage() {
  const navigate = useNavigate();
  const { search } = useLocation();
  const { mailboxEmail = '' } = useParams();
  const selectedEmail = normalizeUiEmail(mailboxEmail);
  const [showSecrets, setShowSecrets] = useState(false);
  const closeDetails = useCallback(() => void navigate(`${mailboxInboxIndexPath()}${persistentMailboxPanelSearch(search)}`), [navigate, search]);
  const closeDeletedMailbox = useCallback((email: string) => {
    if (normalizeUiEmail(email) === selectedEmail) closeDetails();
  }, [closeDetails, selectedEmail]);
  const data = useMailboxData(selectedEmail);
  const actions = useMailboxActions(data, showSecrets, closeDeletedMailbox);
  useMailboxEmailEventCache({ email: data.selected?.email_address, inboxQueryKey: actions.inboxQueryKey, enabled: !!data.selected?.email_address });
  const openDetailTab = useCallback((tab: MailboxDetailTab) => {
    if (data.selected) void navigate(`${mailboxDetailPath(data.selected.email_address, tab)}${persistentMailboxPanelSearch(search)}`);
  }, [data.selected, navigate, search]);

  return (
    <>
      <WorkspacePanel>
        <div className="mailboxWorkspace">
          <div className="mailboxSourcePane">
            <MailboxPageStatus
              total={data.mailboxes.length}
              runningCount={data.runningOperationByEmail.size}
              showSecrets={showSecrets}
              error={data.loadError}
            />
            <MailboxPanel
              mailboxes={data.mailboxes}
              domains={data.domains}
              providerCapabilities={data.providerCapabilities}
              selected={selectedEmail}
              busy={data.busy}
              showSecrets={showSecrets}
              oauthing={actions.oauthing}
              inboxLoading={actions.inboxLoading}
              domainSyncing={actions.domainSyncing}
              runningOperationByEmail={data.runningOperationByEmail}
              hasMoreMailboxes={data.hasMoreMailboxes}
              loadingMoreMailboxes={data.loadingMoreMailboxes}
              onLoadMoreMailboxes={data.loadMoreMailboxes}
              onOAuth={actions.runOAuth}
              onFetchInbox={() => actions.fetchInbox()}
              onSyncDomains={actions.syncProviderDomains}
              onToggleSecrets={() => setShowSecrets((value) => !value)}
              onDelete={actions.requestDeleteMailbox}
              onDone={actions.done}
              onError={actions.toast.showError}
            />
          </div>
          <Outlet context={{ data, actions, showSecrets, closeDetails, openDetailTab } satisfies MailboxPageContext} />
        </div>
      </WorkspacePanel>
      <MailboxDeleteDialog
        mailbox={actions.deleteTarget}
        showSecrets={showSecrets}
        busy={actions.deleting}
        onCancel={actions.cancelDeleteMailbox}
        onConfirm={actions.confirmDeleteMailbox}
      />
    </>
  );
}

export function MailboxIndexRoute() {
  const { data } = useOutletContext<MailboxPageContext>();
  return <MailboxReadingEmpty busy={data.busy} total={data.mailboxes.length} />;
}

export function MailboxOverviewRoute() {
  return <MailboxDetailRoute detailTab="overview" />;
}

export function MailboxInboxRoute() {
  return <MailboxDetailRoute detailTab="inbox" />;
}

function MailboxDetailRoute({ detailTab }: { detailTab: MailboxDetailTab }) {
  const { data, actions, showSecrets, closeDetails, openDetailTab } = useOutletContext<MailboxPageContext>();
  if (!data.selected) {
    return <MailboxReadingEmpty busy={data.busy} total={data.mailboxes.length} />;
  }

  return (
    <MailboxReadingPane
      mailbox={data.selected}
      providerCapability={capabilityForProvider(data.providerCapabilities, data.selected.provider_key)}
      activeTab={detailTab}
      showSecrets={showSecrets}
      inboxResult={actions.inboxResult}
      inboxLoading={actions.inboxLoading}
      canFetchInbox={canRunProviderMailboxAction(
        data.providerCapabilities,
        data.selected,
        MailboxProviderAction.MAILBOX_PROVIDER_ACTION_FETCH_INBOX
      )}
      onClose={closeDetails}
      onTabChange={openDetailTab}
      onCopy={actions.toast.copyValue}
      onFetchInbox={actions.fetchInbox}
      onDelete={actions.requestDeleteMailbox}
    />
  );
}
