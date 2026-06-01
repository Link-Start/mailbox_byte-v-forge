import { Mail } from 'lucide-react';
import { AccountManagementFrame } from '@byte-v-forge/common-ui';
import { MailboxRecordList } from './mailbox-list';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { mailboxProviderConfig } from './mailbox-utils';

export function OutlookMailboxProviderPanel(props: MailboxProviderPanelProps) {
  const config = mailboxProviderConfig('outlook');
  return (
    <AccountManagementFrame title={`${config.label}邮箱账号`} icon={<Mail size={16} />} actions={props.actions}>
      <MailboxRecordList {...props} providerCapability={props.capability} showStatus={config.showStatus} emptyText={`暂无 ${config.label} 邮箱。`} />
    </AccountManagementFrame>
  );
}
