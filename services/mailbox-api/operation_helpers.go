package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"mailboxapi/internal/pagex"
	"mailboxapi/internal/randx"
)

func (s *server) updateOperation(ctx context.Context, operationID string, update operationUpdate) {
	operation, err := s.operations.update(ctx, operationID, update)
	if err != nil {
		log.Printf("update mailbox operation failed operation=%s: %s", operationID, safeMailboxError(err))
		return
	}
	s.hot.PublishOperation(ctx, operation)
}

func normalizedLimit(limit int32) int32 {
	return int32(pagex.NormalizePageLimit(int(limit)))
}

func operationID(prefix string) string {
	if id, err := randx.Hex(8); err == nil {
		return prefix + "-" + id
	}
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
