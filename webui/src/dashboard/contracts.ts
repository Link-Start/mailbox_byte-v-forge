import type { ReactNode } from 'react';

export {
  MailboxAuthStatus,
  MailboxCredentialKind,
  MailboxProviderAction
} from '../proto/byte/v/forge/contracts/mailbox/v1/mailbox';

export type {
  DeleteMailboxResponse,
  EmailInboxMessage,
  EmailMailbox,
  EmailSignal,
  FetchMailboxInboxResult,
  FetchMailboxInboxesRequest,
  FetchMailboxInboxesResponse,
  GetMailboxInboxMessageRequest,
  GetMailboxInboxMessageResponse,
  ListEmailMailboxesResponse,
  ListMailboxDomainsResponse,
  ListMailboxInboxResponse,
  ListMailboxOperationsResponse,
  ListMailboxProviderCapabilitiesResponse,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability,
  MailboxProviderCapabilities,
  StartMailboxOAuthRequest,
  StartMailboxOAuthResponse,
  SyncMailboxDomainsRequest,
  SyncMailboxDomainsResponse,
  UpsertEmailMailboxRequest,
  UpsertEmailMailboxResponse
} from '../proto/byte/v/forge/contracts/mailbox/v1/mailbox';

export enum DashboardNavSection {
  DASHBOARD_NAV_SECTION_INFRASTRUCTURE = 'DASHBOARD_NAV_SECTION_INFRASTRUCTURE'
}

export type DashboardModuleRegistration = {
  manifest: {
    id: string;
    nav: Array<{
      key: string;
      label: string;
      icon: string;
      section: DashboardNavSection;
      required_services?: string[];
      order?: number;
    }>;
  };
  icons: Record<string, ReactNode>;
  views: Record<string, () => ReactNode>;
};
