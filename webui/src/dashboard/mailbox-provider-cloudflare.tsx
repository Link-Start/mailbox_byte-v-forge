import { AccountManagementFrame } from './dashboard-kit';
import { MailboxDomainGroups } from './mailbox-domain-groups';
import { mailboxProviderMatches } from './mailbox-provider-config';
import { providerShowsCredentialState } from './mailbox-provider-capabilities';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';

export function CloudflareMailboxProviderPanel(props: MailboxProviderPanelProps) {
  const providerKey = props.capability?.key || 'cloudflare';
  const configuredDomains = (props.searchQuery ? [] : props.domains)
    .filter((domain) => mailboxProviderMatches(domain.provider_key, providerKey))
    .map((domain) => domain.domain);
  return (
    <AccountManagementFrame actions={props.actions}>
      <MailboxDomainGroups
        {...props}
        providerCapability={props.capability}
        configuredDomains={configuredDomains}
        showStatus={providerShowsCredentialState(props.capability)}
        emptyDomainsText={props.searchQuery ? '无匹配' : '暂无域名'}
        emptyDomainText={props.searchQuery ? '无匹配' : '暂无邮箱'}
      />
    </AccountManagementFrame>
  );
}
