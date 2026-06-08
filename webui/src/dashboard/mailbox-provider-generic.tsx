import { Mail } from 'lucide-react';
import { AccountManagementFrame } from './dashboard-kit';
import { MailboxRecordList } from './mailbox-list';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerDisplayName, providerShowsCredentialState } from './mailbox-utils';

export function GenericMailboxProviderPanel(props: MailboxProviderPanelProps) {
  const label = providerDisplayName(props.capability, props.capability?.key || '邮箱');
  return (
    <AccountManagementFrame title={`${label}邮箱账号`} icon={<Mail size={16} />} actions={props.actions}>
      <MailboxRecordList {...props} providerCapability={props.capability} showStatus={providerShowsCredentialState(props.capability)} emptyText={props.searchQuery ? '没有匹配的邮箱。' : `暂无 ${label} 邮箱。`} />
    </AccountManagementFrame>
  );
}
