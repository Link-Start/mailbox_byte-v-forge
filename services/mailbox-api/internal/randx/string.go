package randx

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

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
