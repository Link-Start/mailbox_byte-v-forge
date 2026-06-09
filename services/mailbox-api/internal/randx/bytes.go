package randx

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

func Bytes(size int) ([]byte, error) {
	if size <= 0 {
		return []byte{}, nil
	}
	out := make([]byte, size)
	if _, err := rand.Read(out); err != nil {
		return nil, err
	}
	return out, nil
}

func Hex(size int) (string, error) {
	raw, err := Bytes(size)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func Base64URL(size int) (string, error) {
	raw, err := Bytes(size)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
