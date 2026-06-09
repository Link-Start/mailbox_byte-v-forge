package inboxapp

import (
	"context"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *Service) latestCachedMessage(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, parserProfile string, signalKind mailboxv1.EmailSignalKind) (*mailboxv1.EmailInboxMessage, bool, error) {
	if s == nil || s.recent == nil {
		return nil, false, nil
	}
	return s.recent.Latest(ctx, email, subjectKeyword, issuedAfterUnix, parserProfile, signalKind)
}

func (s *Service) recordRecentInboxMessages(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) {
	if s == nil || s.recent == nil || len(messages) == 0 {
		return
	}
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if err := s.recent.Record(cacheCtx, messages); err != nil {
		s.log("record recent mailbox email cache failed: %v", err)
	}
}
