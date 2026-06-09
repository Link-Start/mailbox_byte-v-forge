package pagex

import (
	"fmt"
	"strconv"
)

func ParseOffsetToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(pageToken)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("page_token must be a non-negative offset")
	}
	return offset, nil
}

func OffsetToken(offset int) string {
	if offset <= 0 {
		return ""
	}
	return strconv.Itoa(offset)
}
