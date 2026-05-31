package main

import (
	"strings"

	"github.com/byte-v-forge/common-lib/redactx"
)

const mailboxErrorSnippetLimit = 600

func safeMailboxText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}

func safeMailboxError(err error) (value string) {
	if err == nil {
		return ""
	}
	defer func() {
		if recover() != nil {
			value = "mailbox error"
		}
	}()
	return safeMailboxText(err.Error())
}

func safeMailboxErrorMessage(prefix string, err error) string {
	prefix = strings.TrimSpace(prefix)
	message := safeMailboxError(err)
	if prefix == "" {
		return message
	}
	if message == "" {
		return prefix
	}
	return prefix + ": " + message
}

func safeMailboxLogArgs(args ...any) []any {
	out := make([]any, 0, len(args))
	for _, arg := range args {
		switch value := arg.(type) {
		case error:
			out = append(out, safeMailboxError(value))
		case string:
			out = append(out, safeMailboxText(value))
		default:
			out = append(out, arg)
		}
	}
	return out
}
