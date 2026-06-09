package httpx

import (
	"bytes"
	"compress/gzip"
	"io"
)

const DefaultMaxBodyBytes int64 = 8 * 1024 * 1024

func ReadLimited(body io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		limit = DefaultMaxBodyBytes
	}
	return io.ReadAll(io.LimitReader(body, limit))
}

func ReadMaybeGzipLimited(body io.Reader, limit int64) ([]byte, error) {
	raw, err := ReadLimited(body, limit)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(raw, []byte{0x1f, 0x8b}) {
		return raw, nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ReadLimited(reader, limit)
}

func Successful(status int) bool {
	return status >= 200 && status < 300
}
