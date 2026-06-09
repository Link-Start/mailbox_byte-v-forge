package randx

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

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
