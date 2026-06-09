package natseventbus

import "strings"

func (b *Bus) Close() {
	if b == nil || b.conn == nil {
		return
	}
	b.conn.Drain()
	b.conn.Close()
}

func normalizedStreamName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultStream
	}
	return value
}
