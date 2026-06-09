package mailboxmodel

import (
	"errors"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

const (
	AuthStatusAuthorized        = "AUTHORIZED"
	AuthStatusOAuthPending      = "OAUTH_PENDING"
	AuthStatusAuthFailed        = "AUTH_FAILED"
	AuthStatusNeedsManualVerify = "NEEDS_MANUAL_VERIFICATION"
)

var ErrInvalidMailboxListCursor = errors.New("invalid mailbox cursor")

func PublicAuthStatus(value string) mailboxv1.MailboxAuthStatus {
	switch value {
	case AuthStatusOAuthPending:
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_OAUTH_PENDING
	case AuthStatusAuthorized:
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_AUTHORIZED
	case AuthStatusAuthFailed:
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_AUTH_FAILED
	case AuthStatusNeedsManualVerify:
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_NEEDS_MANUAL_VERIFICATION
	case "PASSWORD_ONLY":
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_PASSWORD_ONLY
	case "WEBHOOK_ONLY":
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_WEBHOOK_ONLY
	case "DISABLED":
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_DISABLED
	case "UNKNOWN":
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_UNKNOWN
	default:
		return mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_UNKNOWN
	}
}
