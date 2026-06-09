import type { ReactNode } from 'react';
import type { Mailbox, MailboxDomain, MailboxOperation, MailboxProviderCapability } from './types';

export type MailboxPanelMode = 'inbox' | 'accounts';

export type MailboxProviderPanelProps = {
  mailboxes: Mailbox[];
  mode: MailboxPanelMode;
  domains: MailboxDomain[];
  capability?: MailboxProviderCapability;
  actions?: ReactNode;
  selected?: string;
  busy: boolean;
  showSecrets: boolean;
  oauthing: string;
  inboxLoading: boolean;
  domainSyncing: boolean;
  runningOperationByEmail: Map<string, MailboxOperation>;
  searchQuery?: string;
  hasMoreMailboxes?: boolean;
  loadingMoreMailboxes?: boolean;
  onLoadMoreMailboxes: () => void | Promise<void>;
  onOAuth: (emailAddress?: string) => Promise<void>;
  onFetchInbox: () => Promise<void>;
  onSyncDomains: (providerKey: string) => Promise<void>;
  onToggleSecrets: () => void;
  onDelete: (mailbox: Mailbox) => void;
  onDone: (message: string) => void;
  onError: (message: string) => void;
};
