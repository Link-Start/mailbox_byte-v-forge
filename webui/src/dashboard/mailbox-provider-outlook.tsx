import { MailOutlined } from '@ant-design/icons';
import { AccountManagementFrame } from './dashboard-kit';
import { MailboxRecordList } from './mailbox-list';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerDisplayName, providerShowsCredentialState } from './mailbox-utils';

export function OutlookMailboxProviderPanel(props: MailboxProviderPanelProps) {
  const label = providerDisplayName(props.capability, 'Outlook');
  return (
    <AccountManagementFrame title={`${label}邮箱账号`} icon={<MailOutlined />} actions={props.actions}>
      <MailboxRecordList {...props} providerCapability={props.capability} showStatus={providerShowsCredentialState(props.capability)} emptyText={`暂无 ${label} 邮箱。`} />
    </AccountManagementFrame>
  );
}
