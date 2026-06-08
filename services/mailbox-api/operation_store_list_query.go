package main

import (
	"fmt"
	"strings"

	"mailboxapi/internal/emailx"
)

type operationListQuery struct {
	sql  string
	args []any
}

func newOperationListQuery(filter operationListFilter) operationListQuery {
	conditions := []string{}
	args := []any{}
	if value := operationStatusValue(filter.Status); value != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", appendOperationArg(&args, value)))
	}
	if value := operationActionValue(filter.Action); value != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", appendOperationArg(&args, value)))
	}
	if value := emailx.Normalize(filter.EmailAddress); value != "" {
		conditions = append(conditions, fmt.Sprintf("email_address = $%d", appendOperationArg(&args, value)))
	}
	query := operationSelectSQL()
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	args = append(args, normalizedOperationListLimit(filter.Limit))
	query += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d", len(args))
	return operationListQuery{sql: query, args: args}
}

func appendOperationArg(args *[]any, value any) int {
	*args = append(*args, value)
	return len(*args)
}
