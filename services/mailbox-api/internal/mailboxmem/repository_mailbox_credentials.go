package mailboxmem

import (
	"strings"

	"mailboxapi/internal/mailboxmodel"
)

func mergeCredentials(record *mailboxmodel.Record, input *mailboxmodel.Record, existed bool) {
	password := strings.TrimSpace(input.GetPassword())
	refreshToken := strings.TrimSpace(input.GetRefreshToken())
	accessToken := strings.TrimSpace(input.GetAccessToken())
	authStatus := strings.TrimSpace(input.GetAuthStatus())
	lastError := strings.TrimSpace(input.GetLastError())
	if password != "" || !existed {
		record.Password = password
	}
	if refreshToken != "" || !existed {
		record.RefreshToken = refreshToken
	}
	if accessToken != "" || !existed {
		record.AccessToken = accessToken
	}
	switch {
	case authStatus != "":
		record.AuthStatus = authStatus
	case refreshToken != "":
		record.AuthStatus = mailboxmodel.AuthStatusAuthorized
	case !existed && record.AuthStatus == "":
		record.AuthStatus = mailboxmodel.AuthStatusOAuthPending
	}
	if authStatus != "" || lastError != "" || !existed {
		record.LastError = lastError
	}
}
