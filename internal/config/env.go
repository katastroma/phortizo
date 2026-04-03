//revive:disable:package-comments
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// RequireEnv returns the value of the given environment variable or an error
// if it is not set.
func RequireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return val, nil
}

// StringEnv returns the value of the given environment variable, or the
// fallback if not set.
func StringEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// IntEnv returns the integer value of the given environment variable, or the
// fallback if not set. Returns an error if the value is not a valid integer.
func IntEnv(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return v, nil
}

// DurationEnv returns the duration value of the given environment variable, or
// the fallback if not set. Returns an error if the value is not a valid duration.
func DurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return v, nil
}
