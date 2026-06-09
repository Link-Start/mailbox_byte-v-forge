package inboxapp

import (
	"regexp"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

var (
	emailOTPContextPattern    = regexp.MustCompile(`(?i)(?:verification|security|login|one[- ]?time|otp|code|验证码|安全代码)[^0-9]{0,80}([0-9]{4,8})`)
	emailOTPStandalonePattern = regexp.MustCompile(`(^|[^0-9])([0-9]{6})([^0-9]|$)`)
)

func ExtractEmailOTP(message *mailboxv1.EmailInboxMessage) (string, string) {
	if message == nil {
		return "", ""
	}
	text := strings.Join([]string{
		message.GetSubject(),
		message.GetFromAddress(),
		message.GetBodyPreview(),
	}, "\n")
	if match := emailOTPContextPattern.FindStringSubmatch(text); len(match) >= 2 {
		return NormalizeEmailOTP(match[1]), strings.TrimSpace(match[0])
	}
	if match := emailOTPStandalonePattern.FindStringSubmatch(text); len(match) >= 3 {
		return NormalizeEmailOTP(match[2]), strings.TrimSpace(match[0])
	}
	return "", ""
}

func NormalizeEmailOTP(value string) string {
	return strings.TrimSpace(strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "", "-", "").Replace(value))
}
