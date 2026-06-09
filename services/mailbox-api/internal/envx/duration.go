package envx

import (
	"strconv"
	"time"
)

func DurationSeconds(name string, fallback time.Duration) time.Duration {
	value := String(name)
	if value == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func PositiveDurationSeconds(name string, fallback time.Duration) time.Duration {
	seconds := PositiveInt(name, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}

func NonNegativeDurationSeconds(name string, fallback time.Duration) time.Duration {
	seconds := NonNegativeInt(name, int(fallback/time.Second))
	return time.Duration(seconds) * time.Second
}

func FloatDurationSeconds(name string, fallback float64) time.Duration {
	value := String(name)
	if value == "" {
		return time.Duration(fallback * float64(time.Second))
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 {
		return time.Duration(fallback * float64(time.Second))
	}
	return time.Duration(parsed * float64(time.Second))
}
