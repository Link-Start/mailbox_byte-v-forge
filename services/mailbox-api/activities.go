package main

const (
	emailAuthAuthorized              = "AUTHORIZED"
	emailAuthOAuthPending            = "OAUTH_PENDING"
	emailAuthFailed                  = "AUTH_FAILED"
	emailAuthNeedsManualVerification = "NEEDS_MANUAL_VERIFICATION"
)

type mailboxActivities struct {
	outlookRegistration *outlookRegistrationRunner
	emailBackend        emailBackend
	operations          *operationStore
	hot                 *mailboxHotStream
}
