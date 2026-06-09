import type {
  EmailMailbox as Mailbox,
  EmailInboxMessage,
  EmailSignal,
  FetchMailboxInboxResult,
  FetchMailboxInboxesResponse,
  GetMailboxInboxMessageResponse,
  DeleteMailboxResponse,
  ListEmailMailboxesResponse,
  ListMailboxInboxResponse,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability,
  MailboxProviderCapabilities,
  UpsertEmailMailboxRequest,
  UpsertEmailMailboxResponse
} from './contracts';

export type {
  EmailInboxMessage as InboxMessage,
  EmailSignal,
  FetchMailboxInboxResult as InboxResult,
  FetchMailboxInboxesResponse as InboxResponse,
  GetMailboxInboxMessageResponse as InboxMessageResponse,
  DeleteMailboxResponse,
  ListEmailMailboxesResponse,
  ListMailboxInboxResponse,
  Mailbox,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability,
  UpsertEmailMailboxRequest,
  UpsertEmailMailboxResponse
};

export type MailboxProviderCapability = MailboxProviderCapabilities;

export type LatestOtp = {
  detected: boolean;
  secret_resolvable: boolean;
  ref_id: string;
  subject: string;
  received_at_unix: number;
};
