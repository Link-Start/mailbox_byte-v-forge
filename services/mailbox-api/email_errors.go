package main

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func waitError(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return status.Error(codes.Canceled, "request cancelled")
	}
	return status.Error(codes.Internal, safeMailboxError(err))
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not authorized") ||
		strings.Contains(msg, "no refresh token") ||
		strings.Contains(msg, "AUTH_FAILED")
}
