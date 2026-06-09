package envx

import (
	"fmt"
	"strconv"
)

func Int(name string, fallback int) int {
	value := String(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func IntStrict(name string, fallback int) (int, error) {
	value := String(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}

func PositiveInt(name string, fallback int) int {
	parsed := Int(name, fallback)
	if parsed <= 0 {
		return fallback
	}
	return parsed
}

func NonNegativeInt(name string, fallback int) int {
	parsed := Int(name, fallback)
	if parsed < 0 {
		return fallback
	}
	return parsed
}

func PositiveInt32(name string, fallback int32) int32 {
	parsed := PositiveInt(name, int(fallback))
	return int32(parsed)
}
