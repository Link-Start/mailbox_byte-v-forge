package main

import "strings"

func mailboxAuthStatus(refreshToken string, explicitStatus string) string {
	if status := strings.TrimSpace(explicitStatus); status != "" {
		return status
	}
	if strings.TrimSpace(refreshToken) != "" {
		return emailAuthAuthorized
	}
	return "OAUTH_PENDING"
}

func mailboxOAuthFailureStatus(errorMessage string) string {
	errorText := strings.ToLower(strings.TrimSpace(errorMessage))
	if mailboxOAuthFailureIsRuntime(errorMessage) {
		return emailAuthOAuthPending
	}
	if strings.Contains(errorText, "needs_manual_verification") || strings.Contains(errorText, "account.live.com/abuse") {
		return emailAuthNeedsManualVerification
	}
	return emailAuthFailed
}

func mailboxOAuthFailureIsRuntime(errorMessage string) bool {
	errorText := strings.ToLower(strings.TrimSpace(errorMessage))
	return strings.Contains(errorText, "camoufox") ||
		strings.Contains(errorText, "browserserverimpl.js") ||
		strings.Contains(errorText, "server process terminated unexpectedly") ||
		strings.Contains(errorText, "browser automation") ||
		strings.Contains(errorText, "browser unavailable") ||
		strings.Contains(errorText, "me-token-to-replace")
}
