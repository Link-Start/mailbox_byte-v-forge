import type { EmailMailbox } from '../proto/email';
import type {
  EmailInboxMessage,
  EmailSignal,
  FetchMailboxInboxResult,
  FetchMailboxInboxesResponse,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability,
  MailboxProviderCapabilities
} from '@byte-v-forge/common-ui';

export type {
  EmailInboxMessage as InboxMessage,
  EmailSignal,
  FetchMailboxInboxResult as InboxResult,
  FetchMailboxInboxesResponse as InboxResponse,
  MailboxDomain,
  MailboxOperation,
  MailboxProviderActionCapability
};

export type MailboxProviderCapability = MailboxProviderCapabilities;
export type Mailbox = EmailMailbox;

export type LatestOtp = {
  otp: string;
  subject: string;
  received_at_unix: number;
};
