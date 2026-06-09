package inboxapp

import (
	"html"
	"regexp"
	"strings"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func CompactMessageText(value string, limit int) string {
	text := htmlTagPattern.ReplaceAllString(html.UnescapeString(value), " ")
	text = strings.Join(strings.Fields(strings.ReplaceAll(text, "\u00a0", " ")), " ")
	if limit > 0 && len(text) > limit {
		runes := []rune(text)
		if len(runes) > limit {
			return string(runes[:limit])
		}
	}
	return text
}
