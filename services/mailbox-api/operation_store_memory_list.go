package main

import (
	"context"
	"sort"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *memoryOperationStore) list(ctx context.Context, filter operationListFilter) ([]*mailboxv1.MailboxOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := s.filteredRows(filter)
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].UpdatedAt > rows[j].UpdatedAt
	})
	limit := normalizedOperationListLimit(filter.Limit)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	operations := make([]*mailboxv1.MailboxOperation, 0, len(rows))
	for i := range rows {
		operations = append(operations, operationRowToProto(&rows[i]))
	}
	return operations, nil
}

func (s *memoryOperationStore) filteredRows(filter operationListFilter) []mailboxOperationRow {
	statusValue := operationStatusValue(filter.Status)
	actionValue := operationActionValue(filter.Action)
	emailAddress := emailx.Normalize(filter.EmailAddress)
	s.mu.Lock()
	defer s.mu.Unlock()
	rows := make([]mailboxOperationRow, 0, len(s.rows))
	for _, row := range s.rows {
		if memoryOperationRowMatches(row, statusValue, actionValue, emailAddress) {
			rows = append(rows, row)
		}
	}
	return rows
}

func memoryOperationRowMatches(row mailboxOperationRow, statusValue string, actionValue string, emailAddress string) bool {
	if statusValue != "" && row.Status != statusValue {
		return false
	}
	if actionValue != "" && row.Action != actionValue {
		return false
	}
	return emailAddress == "" || row.EmailAddress == emailAddress
}

func normalizedOperationListLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}
