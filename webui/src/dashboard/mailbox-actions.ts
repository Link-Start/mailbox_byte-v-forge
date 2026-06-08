import { useEffect, useMemo, useState } from 'react';
import {
  activeActionTargets,
  api,
  actionTargetStateKey,
  hasActiveAction,
  short,
  type FetchMailboxInboxesRequest,
  type ListMailboxInboxResponse,
  type StartMailboxOAuthRequest,
  type StartMailboxOAuthResponse,
  type SyncMailboxDomainsRequest,
  type SyncMailboxDomainsResponse,
  useQuery,
  useQueryClient,
  useAsyncActionRunner,
  useToastMessage
} from './dashboard-kit';
import { maskEmail, normalizeUiEmail } from './email-utils';
import { mailboxApiPaths, mailboxInboxURL, mailboxURL } from './mailbox-api-paths';
import type { MailboxData } from './mailbox-data';
import { capabilityForProvider, providerDisplayName } from './mailbox-utils';
import type { DeleteMailboxResponse, InboxResponse, InboxResult, Mailbox } from './types';

export const mailboxInboxQueryKey = (email: string) => ['mailbox', 'inbox', normalizeUiEmail(email)] as const;

export function useMailboxActions(data: MailboxData, showSecrets: boolean, onMailboxDeleted: (email: string) => void) {
  const toast = useToastMessage();
  const queryClient = useQueryClient();
  const selectedEmail = normalizeUiEmail(data.selected?.email_address || '');
  const selectedInboxKey = useMemo(() => mailboxInboxQueryKey(selectedEmail), [selectedEmail]);
  const [deleteTarget, setDeleteTarget] = useState<Mailbox | null>(null);
  const inboxQuery = useQuery<InboxResult | null>({
    queryKey: selectedInboxKey,
    queryFn: () => selectedEmail ? fetchStoredInbox(selectedEmail) : Promise.resolve(null),
    enabled: !!selectedEmail,
    initialData: null
  });
  const runner = useAsyncActionRunner();

  useEffect(() => { if (data.loadError) toast.showError(data.loadError); }, [data.loadError, toast]);

  async function runOAuth(emailAddress = '') {
    const target = emailAddress.trim() || '*';
    await runner.tryRun(actionTargetStateKey('oauth', target), async () => {
      const input: StartMailboxOAuthRequest = { email_address: emailAddress, only_missing: !emailAddress, limit: 100 };
      const resp = await api<StartMailboxOAuthResponse>(mailboxApiPaths.mailboxOAuth, { method: 'POST', body: JSON.stringify(input) });
      toast.showToast(!resp.started || resp.error_message ? 'error' : 'ok', resp.error_message || (!resp.started ? 'OAuth 流程启动失败' : `OAuth 流程已提交: ${short(resp.operation_id)}`));
      await data.invalidate();
    }, { onError: toast.showError });
  }

  async function fetchInbox(emailAddress = '') {
    const target = normalizeUiEmail(emailAddress);
    await runner.tryRun(actionTargetStateKey('fetch-inbox', target || '*'), async () => {
      const input: FetchMailboxInboxesRequest = { limit_per_mailbox: 10, max_mailboxes: target ? 1 : 200, email_address: target, parser_profile: '', received_after_unix: 0 };
      const resp = await api<InboxResponse>(mailboxApiPaths.mailboxInboxFetch, { method: 'POST', body: JSON.stringify(input) });
      for (const result of resp.results || []) {
        const email = result.mailbox?.email_address || result.messages?.[0]?.mailbox_email || target;
        if (email) queryClient.setQueryData(mailboxInboxQueryKey(email), result);
      }
      toast.showToast(resp.failed_count > 0 ? 'error' : 'ok', `${target ? `${showSecrets ? target : maskEmail(target)} ` : ''}收信完成：${resp.message_count} 封邮件`);
      await data.invalidate();
    }, { onError: toast.showError });
  }

  async function syncProviderDomains(providerKey: string) {
    const targetProvider = providerKey.trim();
    if (!targetProvider) {
      toast.showError('provider_key is required');
      return;
    }
    await runner.tryRun(actionTargetStateKey('sync-domains', targetProvider), async () => {
      const input: SyncMailboxDomainsRequest = { provider_key: targetProvider };
      const resp = await api<SyncMailboxDomainsResponse>(mailboxApiPaths.domains, {
        method: 'POST',
        body: JSON.stringify(input)
      });
      const capability = capabilityForProvider(data.providerCapabilities, targetProvider);
      toast.showToast(resp.error_message ? 'error' : 'ok', resp.error_message || `${providerDisplayName(capability, targetProvider)} 域名已同步: ${resp.synced_count || 0}`);
      await data.invalidate();
    }, { onError: toast.showError });
  }

  function requestDeleteMailbox(mailbox: Mailbox) {
    setDeleteTarget(mailbox);
  }

  async function confirmDeleteMailbox() {
    const mailbox = deleteTarget;
    if (!mailbox) return;
    setDeleteTarget(null);
    await runner.tryRun(actionTargetStateKey('delete-mailbox', mailbox.email_address), async () => {
      await api<DeleteMailboxResponse>(mailboxURL(mailbox.email_address), { method: 'DELETE' });
      onMailboxDeleted(mailbox.email_address);
      toast.showOK('邮箱已删除');
      await data.invalidate();
    }, { onError: toast.showError });
  }

  async function done(message: string) {
    toast.showOK(message);
    await data.invalidate();
  }

  return {
    toast,
    deleteTarget,
    inboxResult: inboxQuery.data ?? null,
    inboxQueryKey: selectedInboxKey,
    oauthing: activeActionTargets(runner.activeKeys, 'oauth')[0] || '',
    inboxLoading: inboxQuery.isFetching || hasActiveAction(runner.activeKeys, 'fetch-inbox'),
    domainSyncing: hasActiveAction(runner.activeKeys, 'sync-domains'),
    deleting: hasActiveAction(runner.activeKeys, 'delete-mailbox'),
    runOAuth,
    fetchInbox,
    syncProviderDomains,
    requestDeleteMailbox,
    cancelDeleteMailbox: () => setDeleteTarget(null),
    confirmDeleteMailbox,
    done
  };
}

async function fetchStoredInbox(email: string) {
  const resp = await api<ListMailboxInboxResponse>(mailboxInboxURL(email));
  return resp.result || null;
}
