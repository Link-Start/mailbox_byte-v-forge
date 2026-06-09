import { useCallback } from 'react';
import { Outlet, useLocation, useNavigate, useOutletContext, useParams } from 'react-router';
import { MailboxProviderAction } from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { useMailboxActions } from './mailbox-actions';
import { useMailboxData } from './mailbox-data';
import { MailboxAppNav } from './mailbox-app-nav';
import { MailboxDeleteDialog } from './mailbox-delete-dialog';
import { MailboxPageStatus } from './mailbox-page-status';
import { persistentMailboxPanelSearch } from './mailbox-panel-query';
import { MailboxReadingEmpty, MailboxReadingPane } from './mailbox-reading-pane';
import { mailboxAccountsIndexPath, mailboxInboxIndexPath } from './mailbox-route-paths';
import { MailboxPanel } from './mailboxes';
import { canRunProviderMailboxAction, capabilityForProvider } from './mailbox-provider-capabilities';
import type { MailboxPanelMode } from './mailbox-provider-types';

type MailboxPageContext = {
  data: ReturnType<typeof useMailboxData>;
  actions: ReturnType<typeof useMailboxActions>;
  closeDetails: () => void;
  panelMode: MailboxPanelMode;
};

export function MailboxPage() {
  const navigate = useNavigate();
  const { pathname, search } = useLocation();
  const { mailboxEmail = '' } = useParams();
  const selectedEmail = normalizeUiEmail(mailboxEmail);
  const panelMode: MailboxPanelMode = pathname.split('/').includes('accounts') ? 'accounts' : 'inbox';
  const closeDetails = useCallback(() => {
    const path = panelMode === 'accounts' ? mailboxAccountsIndexPath() : mailboxInboxIndexPath();
    void navigate(`${path}${persistentMailboxPanelSearch(search)}`);
  }, [navigate, panelMode, search]);
  const closeDeletedMailbox = useCallback((email: string) => {
    if (normalizeUiEmail(email) === selectedEmail) closeDetails();
  }, [closeDetails, selectedEmail]);
  const data = useMailboxData(selectedEmail);
  const actions = useMailboxActions(data, closeDeletedMailbox);
  return (
    <>
      <main className="workspacePanel">
        <div className="mailboxAppShell">
          <MailboxAppNav />
          <div className="mailboxWorkspace">
            <div className="mailboxSourcePane">
              <MailboxPageStatus
                total={data.mailboxes.length}
                runningCount={data.runningOperationByEmail.size}
                error={data.loadError}
              />
              <MailboxPanel
                mailboxes={data.mailboxes}
                mode={panelMode}
                providerCapabilities={data.providerCapabilities}
                selected={selectedEmail}
                busy={data.busy}
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
                onDelete={actions.requestDeleteMailbox}
                onDone={actions.done}
                onError={actions.toast.showError}
              />
            </div>
            <Outlet context={{ data, actions, panelMode, closeDetails } satisfies MailboxPageContext} />
          </div>
        </div>
      </main>
      <MailboxDeleteDialog
        mailbox={actions.deleteTarget}
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

export function MailboxAccountsRoute() {
  const { data } = useOutletContext<MailboxPageContext>();
  return <MailboxReadingEmpty busy={data.busy} total={data.mailboxes.length} />;
}

export function MailboxOverviewRoute() {
  return <MailboxDetailRoute />;
}

export function MailboxInboxRoute() {
  return <MailboxDetailRoute />;
}

function MailboxDetailRoute() {
  const { data, actions, panelMode, closeDetails } = useOutletContext<MailboxPageContext>();
  if (!data.selected) {
    return <MailboxReadingEmpty busy={data.busy} total={data.mailboxes.length} />;
  }

  return (
    <MailboxReadingPane
      mailbox={data.selected}
      providerCapability={capabilityForProvider(data.providerCapabilities, data.selected.provider_key)}
      mode={panelMode}
      inboxLoading={actions.inboxLoading}
      canFetchInbox={canRunProviderMailboxAction(
        data.providerCapabilities,
        data.selected,
        MailboxProviderAction.MAILBOX_PROVIDER_ACTION_FETCH_INBOX
      )}
      onClose={closeDetails}
      onCopy={actions.toast.copyValue}
      onFetchInbox={actions.fetchInbox}
      onDelete={actions.requestDeleteMailbox}
    />
  );
}
