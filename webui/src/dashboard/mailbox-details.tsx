import { useEffect, useState } from 'react';
import { DeleteOutlined } from '@ant-design/icons';
import {
  ActionButtonGroup,
  Card,
  ContentTabs,
  KVList,
  MailboxCredentialKind,
  StatusBadge
} from './dashboard-kit';
import type { ActionButtonDescriptor, KVDescriptor } from './dashboard-kit';
import { maskEmail } from './email-utils';
import { mailboxStatusText } from './labels';
import { MailboxInboxSection } from './mailbox-inbox';
import { MailboxOtpPanel } from './otp-panel';
import { latestOtpForInboxResult } from './mailbox-signal-utils';
import { authStatus, mailboxCredentialValue, providerShowsCredentialState, tokenText } from './mailbox-utils';
import type { InboxResult, LatestOtp, Mailbox, MailboxProviderCapability } from './types';

export function MailboxDetails({ mailbox, providerCapability, showSecrets, inboxResult, inboxLoading, canFetchInbox, onCopy, onFetchInbox, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  showSecrets: boolean;
  inboxResult?: InboxResult | null;
  inboxLoading: boolean;
  canFetchInbox: boolean;
  onCopy: (label: string, value: string) => void;
  onFetchInbox: (emailAddress?: string) => Promise<void>;
  onDelete: (mailbox: Mailbox) => Promise<void>;
}) {
  const [activeTab, setActiveTab] = useState<'overview' | 'inbox'>('overview');
  const inboxMessageCount = inboxResult?.messages?.length || 0;
  const latestOtp = latestOtpForInboxResult(inboxResult || null, mailbox.email_address);

  useEffect(() => {
    setActiveTab('overview');
  }, [mailbox.email_address]);

  return (
    <ContentTabs
      value={activeTab}
      onValueChange={(value) => setActiveTab(value as 'overview' | 'inbox')}
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
  onDelete: (mailbox: Mailbox) => Promise<void>;
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
    id: 'password',
    label: '密码',
    ...credentialDisplay(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD, showSecrets),
  }, {
    id: 'oauth',
    label: 'OAuth',
    value: mailboxStatusText(authStatus(mailbox)),
  }, {
    id: 'token',
    label: 'Token',
    value: tokenText(mailbox),
  }, {
    id: 'refresh-token',
    label: 'Refresh',
    ...credentialDisplay(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN, showSecrets),
  }, {
    id: 'access-token',
    label: 'Access',
    ...credentialDisplay(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN, showSecrets),
  });
  fields.push({
    id: 'latest-otp',
    label: '验证码',
    value: latestOtp?.captured ? '已捕获' : '-',
    copyValue: '',
    copyDisabled: true,
    masked: false,
    mono: false,
  });
  const actions: ActionButtonDescriptor[] = [{
    id: 'delete-mailbox',
    label: '删除邮箱',
    icon: <DeleteOutlined />,
    variant: 'destructive',
    onClick: () => void onDelete(mailbox),
  }];

  return (
    <section className="grid gap-3">
      <Card className="grid gap-2 p-3 shadow-none">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <strong className="block truncate text-sm">{showSecrets ? mailbox.email_address : maskEmail(mailbox.email_address)}</strong>
          </div>
          {showCredentialState && <StatusBadge status={authStatus(mailbox)} />}
        </div>
        <MailboxOtpPanel latestOtp={latestOtp} showSecrets={showSecrets} loading={false} onCopy={onCopy} />
      </Card>
      <div>
        <KVList items={fields} onCopy={onCopy} />
      </div>
      <ActionButtonGroup actions={actions} />
    </section>
  );
}

function credentialDisplay(mailbox: Mailbox, kind: MailboxCredentialKind, showSecrets: boolean): Pick<KVDescriptor, 'value' | 'copyValue' | 'copyDisabled' | 'masked' | 'mono'> {
  const value = mailboxCredentialValue(mailbox, kind);
  return {
    value: value || '-',
    copyValue: value,
    copyDisabled: !showSecrets || !value,
    masked: !!value && !showSecrets,
    mono: true,
  };
}
