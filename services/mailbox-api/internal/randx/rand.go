package randx

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	AlphabetAlphaLower = "abcdefghijklmnopqrstuvwxyz"
	AlphabetAlphaNum   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	AlphabetLowerNum   = "abcdefghijklmnopqrstuvwxyz0123456789"
	AlphabetURLSafe    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
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

func String(alphabet string, length int) (string, error) {
	alphabet = strings.TrimSpace(alphabet)
	if length <= 0 {
		return "", nil
	}
	if alphabet == "" {
		return "", fmt.Errorf("alphabet is empty")
	}
	max := big.NewInt(int64(len(alphabet)))
	out := make([]byte, length)
	for idx := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[idx] = alphabet[n.Int64()]
	}
	return string(out), nil
}

func Index(size int) (int, error) {
	if size <= 0 {
		return 0, fmt.Errorf("size must be positive")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(size)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func Int(maxExclusive int64) (int64, error) {
	if maxExclusive <= 0 {
		return 0, fmt.Errorf("maxExclusive must be positive")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(maxExclusive))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

func PositiveInt63() (int64, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 63)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	if n.Sign() <= 0 {
		return 0, fmt.Errorf("generated value is not positive")
	}
	return n.Int64(), nil
}

func IntRange(minValue int, maxValue int) (int, error) {
	if maxValue <= minValue {
		return minValue, nil
	}
	n, err := Int(int64(maxValue - minValue + 1))
	if err != nil {
		return minValue, err
	}
	return minValue + int(n), nil
}
