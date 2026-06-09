import { useEffect, useMemo, useState } from 'react';
import {
  activeActionTargets,
  actionTargetStateKey,
  hasActiveAction,
  short,
  useQuery,
  useQueryClient,
  useAsyncActionRunner,
  useToastMessage
} from './dashboard-kit';
import { normalizeUiEmail } from './email-utils';
import { deleteMailbox, fetchMailboxInboxes, fetchStoredInbox, startMailboxOAuth, syncMailboxDomains } from './mailbox-action-api';
import type { MailboxData } from './mailbox-data';
import { capabilityForProvider, providerDisplayName } from './mailbox-provider-capabilities';
import type { InboxResult, Mailbox } from './types';

export const mailboxInboxQueryKey = (email: string) => ['mailbox', 'inbox', normalizeUiEmail(email)] as const;

export function useMailboxActions(data: MailboxData, onMailboxDeleted: (email: string) => void, options: { loadInbox?: boolean } = {}) {
  const toast = useToastMessage();
  const queryClient = useQueryClient();
  const selectedEmail = normalizeUiEmail(data.selected?.email_address || '');
  const selectedInboxKey = useMemo(() => mailboxInboxQueryKey(selectedEmail), [selectedEmail]);
  const [deleteTarget, setDeleteTarget] = useState<Mailbox | null>(null);
  const inboxQuery = useQuery<InboxResult | null>({
    queryKey: selectedInboxKey,
    queryFn: () => selectedEmail ? fetchStoredInbox(selectedEmail) : Promise.resolve(null),
    enabled: !!selectedEmail && options.loadInbox !== false,
    initialData: null
  });
  const runner = useAsyncActionRunner();

  useEffect(() => { if (data.loadError) toast.showError(data.loadError); }, [data.loadError, toast]);

  async function runOAuth(emailAddress = '') {
    const target = emailAddress.trim() || '*';
    await runner.tryRun(actionTargetStateKey('oauth', target), async () => {
      const resp = await startMailboxOAuth(emailAddress);
      toast.showToast(!resp.started || resp.error_message ? 'error' : 'ok', resp.error_message || (!resp.started ? 'OAuth 流程启动失败' : `OAuth 流程已提交: ${short(resp.operation_id)}`));
      await data.invalidate();
    }, { onError: toast.showError });
  }

  async function fetchInbox(emailAddress = '') {
    const target = normalizeUiEmail(emailAddress);
    await runner.tryRun(actionTargetStateKey('fetch-inbox', target || '*'), async () => {
      const resp = await fetchMailboxInboxes(target);
      for (const result of resp.results || []) {
        const email = result.mailbox?.email_address || result.messages?.[0]?.mailbox_email || target;
        if (email) {
          queryClient.setQueryData(mailboxInboxQueryKey(email), result);
          await queryClient.invalidateQueries({ queryKey: mailboxInboxQueryKey(email) });
        }
      }
      toast.showToast(resp.failed_count > 0 ? 'error' : 'ok', `${target ? `${target} ` : ''}收信完成：${resp.message_count} 封邮件`);
      if (resp.message_count > 0 || resp.failed_count > 0) await data.invalidate();
    }, { onError: toast.showError });
  }

  async function syncProviderDomains(providerKey: string) {
    const targetProvider = providerKey.trim();
    if (!targetProvider) {
      toast.showError('provider_key is required');
      return;
    }
    await runner.tryRun(actionTargetStateKey('sync-domains', targetProvider), async () => {
      const resp = await syncMailboxDomains(targetProvider);
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
      await deleteMailbox(mailbox.email_address);
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
