import { Mail } from 'lucide-react';
import { AccountManagementFrame } from './dashboard-kit';
import { MailboxDomainGroups } from './mailbox-list';
import { mailboxProviderMatches, providerDisplayName, providerShowsCredentialState } from './mailbox-utils';
import type { MailboxProviderPanelProps } from './mailbox-provider-types';

export function CloudflareMailboxProviderPanel(props: MailboxProviderPanelProps) {
  const providerKey = props.capability?.key || 'cloudflare';
  const label = providerDisplayName(props.capability, 'Cloudflare');
  const configuredDomains = (props.searchQuery ? [] : props.domains)
    .filter((domain) => mailboxProviderMatches(domain.provider_key, providerKey))
    .map((domain) => domain.domain);
  return (
    <AccountManagementFrame title={`${label}邮箱账号`} icon={<Mail size={16} />} actions={props.actions}>
      <MailboxDomainGroups
        {...props}
        providerCapability={props.capability}
        configuredDomains={configuredDomains}
        showStatus={providerShowsCredentialState(props.capability)}
        emptyDomainsText={props.searchQuery ? '没有匹配的邮箱。' : '域名未配置；邮件到达后会按 recipient 自动归组。'}
        emptyDomainText={props.searchQuery ? '这个域名下没有匹配的邮箱。' : '这个域名下暂无邮件地址，收到邮件后会自动出现。'}
      />
    </AccountManagementFrame>
  );
}
