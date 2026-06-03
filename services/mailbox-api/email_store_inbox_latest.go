package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

func (s *MailboxStore) LatestMessage(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64) (*mailboxv1.EmailInboxMessage, bool, error) {
	return s.LatestMessageWithSignal(ctx, email, subjectKeyword, issuedAfterUnix, "", mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED)
}

func (s *MailboxStore) LatestMessageWithSignal(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	candidates := uniqueStrings([]string{email, emailx.CanonicalPlusAlias(email)})
	for _, candidate := range candidates {
		msg, ok, err := s.latestCachedMessage(ctx, candidate, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
		if err != nil {
			logWarning("latest mailbox email cache read failed email=%s: %v", emailx.Redact(candidate), err)
			break
		}
		if ok {
			if err := s.attachEmailSignalSecrets(ctx, msg); err != nil {
				logWarning("latest mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(candidate), err)
			}
			return msg, true, nil
		}
	}
	for _, candidate := range candidates {
		msg, ok, err := s.latestMessageForMailbox(ctx, candidate, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
		if err != nil || ok {
			return msg, ok, err
		}
	}
	return nil, false, nil
}

func (s *MailboxStore) latestCachedMessage(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	if s == nil || s.recent == nil {
		return nil, false, nil
	}
	return s.recent.Latest(ctx, email, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
}

func (s *MailboxStore) recordRecentInboxMessages(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) {
	if s == nil || s.recent == nil || len(messages) == 0 {
		return
	}
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if err := s.recent.Record(cacheCtx, messages); err != nil {
		logWarning("record recent mailbox email cache failed: %v", err)
	}
}

func (s *MailboxStore) latestMessageForMailbox(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, false, nil
	}
	args := []any{email}
	query := inboxMessageSelectSQL + "WHERE mailbox_email = $1"
	if issuedAfterUnix > 0 {
		args = append(args, issuedAfterUnix)
		query += fmt.Sprintf(" AND received_at >= $%d", len(args))
	}
	if keyword := strings.TrimSpace(subjectKeyword); keyword != "" {
		args = append(args, "%"+keyword+"%")
		query += fmt.Sprintf(" AND (subject ILIKE $%d OR body_preview ILIKE $%d OR body_text ILIKE $%d)", len(args), len(args), len(args))
	}
	args = append(args, 50)
	query += fmt.Sprintf(" ORDER BY received_at DESC, updated_at DESC, message_key DESC LIMIT $%d", len(args))

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		row, err := scanInboxMessageRow(rows)
		if err != nil {
			return nil, false, err
		}
		msg, err := row.toProtoForProfile(parserProfile)
		if err != nil {
			return nil, false, err
		}
		if err := s.attachEmailSignalSecrets(ctx, msg); err != nil {
			logWarning("latest mailbox email signal secret refresh failed email=%s: %v", emailx.Redact(email), err)
		}
		if messageHasSignal(msg, signalKind) {
			return msg, true, nil
		}
	}
	return nil, false, rows.Err()
}
