import { Trash2 } from 'lucide-react';
import {
  ActionButtonGroup,
  KVList,
  MailboxCredentialKind,
} from './dashboard-kit';
import type { ActionButtonDescriptor, KVDescriptor } from './dashboard-kit';
import { mailboxStatusText } from './labels';
import { authStatus } from './mailbox-auth-status';
import { mailboxCredentialPresent } from './mailbox-credentials';
import { providerShowsCredentialState } from './mailbox-provider-capabilities';
import { MailboxOtpPanel } from './otp-panel';
import type { LatestOtp, Mailbox, MailboxProviderCapability } from './types';

export function MailboxOverview({ mailbox, providerCapability, latestOtp, onCopy, onDelete }: {
  mailbox: Mailbox;
  providerCapability?: MailboxProviderCapability;
  latestOtp: LatestOtp | null;
  onCopy: (label: string, value: string) => void;
  onDelete: (mailbox: Mailbox) => void;
}) {
  const showCredentialState = providerShowsCredentialState(providerCapability);
  const fields = overviewFields(mailbox, showCredentialState);
  const actions: ActionButtonDescriptor[] = [{
    id: 'delete-mailbox',
    label: '删除邮箱',
    icon: <Trash2 />,
    variant: 'destructive',
    onClick: () => void onDelete(mailbox),
  }];

  return (
    <section className="grid gap-3">
      <MailboxOtpPanel latestOtp={latestOtp} loading={false} />
      <div>
        <KVList items={fields} onCopy={onCopy} />
      </div>
      <ActionButtonGroup actions={actions} />
    </section>
  );
}

function overviewFields(mailbox: Mailbox, showCredentialState: boolean) {
  const fields: KVDescriptor[] = [{
    id: 'email',
    label: '邮箱',
    value: mailbox.email_address,
    copyValue: mailbox.email_address,
    copyDisabled: !mailbox.email_address,
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
    mono: false,
  });
  return fields;
}

function credentialSummary(mailbox: Mailbox) {
  const labels: string[] = [];
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_PASSWORD)) labels.push('密码');
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN)) labels.push('Refresh');
  if (mailboxCredentialPresent(mailbox, MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN)) labels.push('Access');
  return labels.join(' / ');
}
