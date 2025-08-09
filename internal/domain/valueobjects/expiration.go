package valueobjects

import (
	"errors"
	"fmt"
	"time"
)

const (
	DefaultExpirationDuration = 24 * time.Hour
	MinExpirationDuration     = 1 * time.Minute
	MaxExpirationDuration     = 7 * 24 * time.Hour // 7 days
)

var (
	ErrInvalidExpiration = errors.New("invalid expiration time")
	ErrExpired           = errors.New("content has expired")
)

type ExpirationTime struct {
	value time.Time
}

func NewExpirationTime(duration time.Duration) ExpirationTime {
	if duration < MinExpirationDuration {
		duration = MinExpirationDuration
	}
	if duration > MaxExpirationDuration {
		duration = MaxExpirationDuration
	}
	return ExpirationTime{
		value: time.Now().UTC().Add(duration),
	}
}

func NewDefaultExpirationTime() ExpirationTime {
	return NewExpirationTime(DefaultExpirationDuration)
}

func ExpirationTimeFrom(t time.Time) ExpirationTime {
	return ExpirationTime{value: t.UTC()}
}

func (e ExpirationTime) Value() time.Time {
	return e.value
}

func (e ExpirationTime) IsExpired() bool {
	return time.Now().UTC().After(e.value)
}

func (e ExpirationTime) TimeRemaining() time.Duration {
	remaining := e.value.Sub(time.Now().UTC())
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (e ExpirationTime) String() string {
	return e.value.Format(time.RFC3339)
}

func (e ExpirationTime) HumanReadable() string {
	remaining := e.TimeRemaining()
	if remaining == 0 {
		return "Expired"
	}
	
	hours := int(remaining.Hours())
	if hours >= 24 {
		days := hours / 24
		if days == 1 {
			return "Expires in 1 day"
		}
		return fmt.Sprintf("Expires in %d days", days)
	}
	
	if hours > 0 {
		if hours == 1 {
			return "Expires in 1 hour"
		}
		return fmt.Sprintf("Expires in %d hours", hours)
	}
	
	minutes := int(remaining.Minutes())
	if minutes == 1 {
		return "Expires in 1 minute"
	}
	return fmt.Sprintf("Expires in %d minutes", minutes)
}

func (e ExpirationTime) ExtendBy(duration time.Duration) ExpirationTime {
	newExpiration := e.value.Add(duration)
	maxExpiration := time.Now().UTC().Add(MaxExpirationDuration)
	
	if newExpiration.After(maxExpiration) {
		newExpiration = maxExpiration
	}
	
	return ExpirationTime{value: newExpiration}
}