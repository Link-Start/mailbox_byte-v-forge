package main

import (
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func validateOutlookPollableMailbox(row mailboxprovider.MailboxRecord) error {
	if strings.TrimSpace(row.RefreshToken) == "" {
		return fmt.Errorf("mailbox has no refresh token: %s", emailx.Redact(row.Email))
	}
	if row.AuthStatus != mailboxmodel.AuthStatusAuthorized {
		return fmt.Errorf("mailbox is not authorized: %s auth_status=%s", emailx.Redact(row.Email), row.AuthStatus)
	}
	return nil
}
