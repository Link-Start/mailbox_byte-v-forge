package main

import "mailboxapi/internal/mailboxpg"

const (
	emailAuthAuthorized              = "AUTHORIZED"
	emailAuthOAuthPending            = "OAUTH_PENDING"
	emailAuthFailed                  = "AUTH_FAILED"
	emailAuthNeedsManualVerification = "NEEDS_MANUAL_VERIFICATION"
)

type mailboxActivities struct {
	providerActions *mailboxProviderActionRegistry
	emailBackend    emailBackend
	mailboxRepo     *mailboxpg.Repository
	operations      *operationStore
	hot             *mailboxHotStream
}
