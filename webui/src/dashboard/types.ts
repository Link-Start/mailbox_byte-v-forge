import type {
  EmailMailbox as Mailbox,
  EmailInboxMessage,
  EmailSignal,
  FetchMailboxInboxResult,
  FetchMailboxInboxesResponse,
  DeleteMailboxResponse,
  ListEmailMailboxesResponse,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability,
  MailboxProviderCapabilities,
  UpsertEmailMailboxRequest,
  UpsertEmailMailboxResponse
} from '@byte-v-forge/common-ui';

export type {
  EmailInboxMessage as InboxMessage,
  EmailSignal,
  FetchMailboxInboxResult as InboxResult,
  FetchMailboxInboxesResponse as InboxResponse,
  DeleteMailboxResponse,
  ListEmailMailboxesResponse,
  Mailbox,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability,
  UpsertEmailMailboxRequest,
  UpsertEmailMailboxResponse
};

export type MailboxProviderCapability = MailboxProviderCapabilities;

export type LatestOtp = {
  captured: boolean;
  ref_id: string;
  subject: string;
  received_at_unix: number;
};
