import { MailboxCredentialKind } from './dashboard-kit';
import type { Mailbox } from './types';

export function requiredCredentialsPresent(mailbox: Mailbox, credentials: MailboxCredentialKind[]) {
  return credentials.every((credential) => mailboxCredentialPresent(mailbox, credential));
}

export function mailboxCredentialPresent(mailbox: Mailbox, credential: MailboxCredentialKind) {
  switch (credential) {
    case MailboxCredentialKind.MAILBOX_CREDENTIAL_KIND_UNSPECIFIED:
      return true;
    default:
      return (mailbox.credential_state?.present_credentials || []).includes(credential);
  }
}
