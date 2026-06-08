import { AccountManagementFrame } from './dashboard-kit';
import { MailboxRecordList } from './mailbox-list';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';
import { providerShowsCredentialState } from './mailbox-utils';

export function GenericMailboxProviderPanel(props: MailboxProviderPanelProps) {
  return (
    <AccountManagementFrame actions={props.actions}>
      <MailboxRecordList {...props} providerCapability={props.capability} showStatus={providerShowsCredentialState(props.capability)} emptyText={props.searchQuery ? '无匹配' : '暂无邮箱'} />
    </AccountManagementFrame>
  );
}
