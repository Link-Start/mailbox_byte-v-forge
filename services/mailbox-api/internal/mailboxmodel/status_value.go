package mailboxmodel

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

func AuthStatusValue(status mailboxv1.MailboxAuthStatus) string {
	switch status {
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_OAUTH_PENDING:
		return AuthStatusOAuthPending
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_AUTHORIZED:
		return AuthStatusAuthorized
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_AUTH_FAILED:
		return AuthStatusAuthFailed
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_NEEDS_MANUAL_VERIFICATION:
		return AuthStatusNeedsManualVerify
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_PASSWORD_ONLY:
		return "PASSWORD_ONLY"
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_WEBHOOK_ONLY:
		return "WEBHOOK_ONLY"
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_DISABLED:
		return "DISABLED"
	case mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_UNKNOWN:
		return "UNKNOWN"
	default:
		return ""
	}
}
