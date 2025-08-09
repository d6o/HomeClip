package valueobjects

import (
	"strings"
	"testing"
	"time"
)

func TestNewExpirationTime(t *testing.T) {
	tests := []struct {
		name             string
		duration         time.Duration
		expectedMinHours float64
		expectedMaxHours float64
	}{
		{
			name:             "normal duration",
			duration:         24 * time.Hour,
			expectedMinHours: 23.9,
			expectedMaxHours: 24.1,
		},
		{
			name:             "below minimum",
			duration:         30 * time.Second,
			expectedMinHours: 0.01, // 1 minute minimum
			expectedMaxHours: 0.02,
		},
		{
			name:             "above maximum",
			duration:         10 * 24 * time.Hour,
			expectedMinHours: 167.9, // 7 days maximum
			expectedMaxHours: 168.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiration := NewExpirationTime(tt.duration)
			remaining := expiration.TimeRemaining()
			hours := remaining.Hours()
			
			if hours < tt.expectedMinHours || hours > tt.expectedMaxHours {
				t.Errorf("Expected between %v and %v hours, got %v", 
					tt.expectedMinHours, tt.expectedMaxHours, hours)
			}
		})
	}
}

func TestNewDefaultExpirationTime(t *testing.T) {
	expiration := NewDefaultExpirationTime()
	remaining := expiration.TimeRemaining()
	
	// Should be approximately 24 hours
	hours := remaining.Hours()
	if hours < 23.9 || hours > 24.1 {
		t.Errorf("Expected approximately 24 hours, got %v", hours)
	}
}

func TestExpirationTime_IsExpired(t *testing.T) {
	// Create an expiration time that has already passed
	pastExpiration := ExpirationTimeFrom(time.Now().Add(-1 * time.Hour))
	if !pastExpiration.IsExpired() {
		t.Error("Expected past expiration to be expired")
	}
	
	// Create an expiration time in the future
	futureExpiration := ExpirationTimeFrom(time.Now().Add(1 * time.Hour))
	if futureExpiration.IsExpired() {
		t.Error("Expected future expiration to not be expired")
	}
}

func TestExpirationTime_TimeRemaining(t *testing.T) {
	// Test expired time
	pastExpiration := ExpirationTimeFrom(time.Now().Add(-1 * time.Hour))
	if pastExpiration.TimeRemaining() != 0 {
		t.Error("Expected 0 time remaining for expired content")
	}
	
	// Test future time
	futureTime := time.Now().Add(2 * time.Hour)
	futureExpiration := ExpirationTimeFrom(futureTime)
	remaining := futureExpiration.TimeRemaining()
	
	// Should be approximately 2 hours
	if remaining < 1*time.Hour+50*time.Minute || remaining > 2*time.Hour+10*time.Minute {
		t.Errorf("Expected approximately 2 hours remaining, got %v", remaining)
	}
}

func TestExpirationTime_HumanReadable(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{
			name:     "expired",
			duration: -1 * time.Hour,
			contains: "Expired",
		},
		{
			name:     "minutes",
			duration: 30 * time.Minute,
			contains: "minutes",
		},
		{
			name:     "one hour",
			duration: 1*time.Hour + 30*time.Minute,
			contains: "1 hour",
		},
		{
			name:     "multiple hours",
			duration: 5 * time.Hour,
			contains: "5 hours",
		},
		{
			name:     "one day",
			duration: 25 * time.Hour,
			contains: "1 day",
		},
		{
			name:     "multiple days",
			duration: 3 * 24 * time.Hour,
			contains: "3 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiration := ExpirationTimeFrom(time.Now().Add(tt.duration))
			readable := expiration.HumanReadable()
			
			if !strings.Contains(readable, tt.contains) {
				t.Errorf("Expected human readable to contain '%s', got '%s'", 
					tt.contains, readable)
			}
		})
	}
}

func TestExpirationTime_ExtendBy(t *testing.T) {
	// Create an expiration 1 hour from now
	original := ExpirationTimeFrom(time.Now().Add(1 * time.Hour))
	
	// Extend by 1 hour
	extended := original.ExtendBy(1 * time.Hour)
	
	// Should be approximately 2 hours from now
	remaining := extended.TimeRemaining()
	if remaining < 1*time.Hour+50*time.Minute || remaining > 2*time.Hour+10*time.Minute {
		t.Errorf("Expected approximately 2 hours after extension, got %v", remaining)
	}
	
	// Test max extension limit
	veryExtended := original.ExtendBy(10 * 24 * time.Hour)
	maxRemaining := veryExtended.TimeRemaining()
	
	// Should not exceed 7 days
	if maxRemaining > 7*24*time.Hour+1*time.Hour {
		t.Errorf("Extension should not exceed 7 days, got %v", maxRemaining)
	}
}

func TestExpirationTime_String(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	expiration := ExpirationTimeFrom(fixedTime)
	
	str := expiration.String()
	if str != "2024-01-01T12:00:00Z" {
		t.Errorf("Expected RFC3339 format, got %v", str)
	}
}