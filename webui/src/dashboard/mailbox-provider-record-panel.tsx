import { AccountManagementFrame } from './dashboard-kit';
import { MailboxRecordList } from './mailbox-list';
import { providerShowsCredentialState } from './mailbox-provider-capabilities';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';

export function MailboxProviderRecordPanel(props: MailboxProviderPanelProps) {
  return (
    <AccountManagementFrame actions={props.actions}>
      <MailboxRecordList
        {...props}
        providerCapability={props.capability}
        showStatus={providerShowsCredentialState(props.capability)}
        emptyText={props.searchQuery ? '无匹配邮箱' : '暂无邮箱'}
      />
    </AccountManagementFrame>
  );
}
