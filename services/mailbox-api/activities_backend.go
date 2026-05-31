package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/pb"
)

func (a *mailboxActivities) upsertMailbox(ctx context.Context, mailbox *pb.EmailMailbox) error {
	resp, err := a.emailBackend.UpsertMailbox(ctx, &pb.UpsertEmailMailboxRequest{Mailbox: mailbox})
	if err != nil {
		return fmt.Errorf("upsert mailbox %s: %s", emailx.Redact(mailbox.GetEmailAddress()), safeMailboxError(err))
	}
	if resp == nil || resp.GetMailbox() == nil {
		return fmt.Errorf("email service returned empty mailbox for %s", emailx.Redact(mailbox.GetEmailAddress()))
	}
	return nil
}

func (a *mailboxActivities) markEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) error {
	resp, err := a.emailBackend.MarkEmailAuthStatus(ctx, &pb.MarkEmailAuthStatusRequest{
		EmailAddress: emailx.Normalize(email),
		AuthStatus:   authStatus,
		LastError:    strings.TrimSpace(lastError),
	})
	if err != nil {
		return fmt.Errorf("mark mailbox auth status %s: %s", emailx.Redact(email), safeMailboxError(err))
	}
	if resp == nil || resp.GetMailbox() == nil {
		return fmt.Errorf("email service returned empty auth status response for %s", emailx.Redact(email))
	}
	return nil
}
