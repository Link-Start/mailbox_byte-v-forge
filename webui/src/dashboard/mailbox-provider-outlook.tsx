import { MailboxProviderRecordPanel } from './mailbox-provider-record-panel';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';

export function OutlookMailboxProviderPanel(props: MailboxProviderPanelProps) {
  return <MailboxProviderRecordPanel {...props} />;
}
