import { MailboxProviderRecordPanel } from './mailbox-provider-record-panel';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';

export function CloudflareMailboxProviderPanel(props: MailboxProviderPanelProps) {
  return <MailboxProviderRecordPanel {...props} />;
}
