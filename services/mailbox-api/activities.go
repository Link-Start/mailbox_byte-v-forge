package main

const (
	emailAuthAuthorized              = "AUTHORIZED"
	emailAuthOAuthPending            = "OAUTH_PENDING"
	emailAuthFailed                  = "AUTH_FAILED"
	emailAuthNeedsManualVerification = "NEEDS_MANUAL_VERIFICATION"
)

type mailboxActivities struct {
	providerActions *mailboxProviderActionRegistry
	emailBackend    emailBackend
	mailboxRepo     mailboxRepository
	operations      operationStore
	hot             *mailboxHotStream
}
