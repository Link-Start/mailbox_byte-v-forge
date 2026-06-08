import { Trash2 } from 'lucide-react';
import {
  ActionButtonGroup,
  ContentTabs,
  KVList,
  MailboxCredentialKind,
} from './dashboard-kit';
import type { ActionButtonDescriptor, KVDescriptor } from './dashboard-kit';
import { maskEmail } from './email-utils';
import { mailboxStatusText } from './labels';
import { MailboxInboxSection } from './mailbox-inbox';
import { MailboxOtpPanel } from './otp-panel';
import type { MailboxDetailTab } from './mailbox-route-paths';
import { latestOtpForInboxResult } from './mailbox-signal-utils';
import { authStatus } from './mailbox-auth-status';
import { mailboxCredentialPresent } from './mailbox-credentials';
import { providerShowsCredentialState } from './mailbox-provider-capabilities';
import type { InboxResult, LatestOtp, Mailbox, MailboxProviderCapability } from './types';

export function MailboxDetails({ mailbox, providerCapability, activeTab, showSecrets, inboxResult, inboxLoading, canFetchInbox, onTabChange, onCopy, onFetchInbox, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  activeTab: MailboxDetailTab;
  showSecrets: boolean;
  inboxResult?: InboxResult | null;
  inboxLoading: boolean;
  canFetchInbox: boolean;
  onTabChange: (tab: MailboxDetailTab) => void;
  onCopy: (label: string, value: string) => void;
  onFetchInbox: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const inboxMessageCount = inboxResult?.messages?.length || 0;
  const latestOtp = latestOtpForInboxResult(inboxResult || null, mailbox.email_address);

  return (
    <ContentTabs
      value={activeTab}
      onValueChange={(value) => onTabChange(value as MailboxDetailTab)}
      tabsClassName="min-h-0 flex-1 overflow-hidden"
      tabsListVariant="line"
      tabsListClassName="w-full"
      tabs={[
        {
          value: 'overview',
          label: '概览',
          contentClassName: 'overflow-auto',
          content: <MailboxOverview mailbox={mailbox} providerCapability={providerCapability} showSecrets={showSecrets} latestOtp={latestOtp} onCopy={onCopy} onDelete={onDelete} />
        },
        {
          value: 'inbox',
          label: `收件箱 ${inboxMessageCount}`,
          contentClassName: 'overflow-auto',
          content: <MailboxInboxSection mailbox={mailbox} result={inboxResult} showSecrets={showSecrets} loading={inboxLoading} canFetch={canFetchInbox} onFetch={onFetchInbox} />
        }
      ]}
    />
  );
}

function MailboxOverview({ mailbox, providerCapability, showSecrets, latestOtp, onCopy, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  showSecrets: boolean;
  latestOtp: LatestOtp | null;
  onCopy: (label: string, value: string) => void;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const showCredentialState = providerShowsCredentialState(providerCapability);
  const fields: KVDescriptor[] = [{
    id: 'email',
    label: '邮箱',
    value: showSecrets ? mailbox.email_address : maskEmail(mailbox.email_address),
    copyValue: mailbox.email_address,
    copyDisabled: !mailbox.email_address,
    masked: !showSecrets,
  }];
  if (showCredentialState) fields.push({
    id: 'oauth',
    label: '授权',
    value: mailboxStatusText(authStatus(mailbox)),
  });
  const credentials = credentialSummary(mailbox);
  if (showCredentialState && credentials) fields.push({
    id: 'credentials',
    label: '凭据',
    value: credentials,
    copyValue: '',
    copyDisabled: true,
    masked: false,
    mono: false,
  });
  const actions: ActionButtonDescriptor[] = [{
    id: 'delete-mailbox',
    label: '删除邮箱',
    icon: <Trash2 />,
    variant: 'destructive',
    onClick: () => void onDelete(mailbox),
  }];

  return (
    <section className="grid gap-3">
      <MailboxOtpPanel latestOtp={latestOtp} showSecrets={showSecrets} loading={false} />
      <div>
        <KVList items={fields} onCopy={onCopy} />
      </div>
      <ActionButtonGroup actions={actions} />
    </section>
  );
}

function credentialSummary(mailbox: Mailbox) {
  const labels: string[] = [];
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD)) labels.push('密码');
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN)) labels.push('Refresh');
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN)) labels.push('Access');
  return labels.join(' / ');
}
