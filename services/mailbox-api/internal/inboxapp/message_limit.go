package inboxapp

import "time"

func MessageLimitValue(limit int32, fallback int) int {
	n := int(limit)
	if n <= 0 {
		n = fallback
	}
	if n <= 0 {
		n = 25
	}
	if n > 100 {
		n = 100
	}
	return n
}

func InboxReceivedAfter(watermarkNs int64, overlapSeconds int) int64 {
	if watermarkNs <= 0 {
		return 0
	}
	after := watermarkNs - int64(overlapSeconds)*int64(time.Second)
	if after < 0 {
		return 0
	}
	return after
}
