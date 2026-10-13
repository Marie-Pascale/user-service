package domain

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	t.Run("creates a valid session", func(t *testing.T) {
		s, plainToken, err := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plainToken == "" {
			t.Error("the plaintext token should not be empty")
		}
		if s.TokenHash == plainToken {
			t.Error("TokenHash should never equal the plaintext token")
		}
		if !s.ExpiresAt.After(s.CreatedAt) {
			t.Error("ExpiresAt should be after CreatedAt")
		}
		if s.IPAddress != "203.0.113.1" {
			t.Errorf("IPAddress = %q, want %q", s.IPAddress, "203.0.113.1")
		}
		if s.UserAgent != "Mozilla/5.0" {
			t.Errorf("UserAgent = %q, want %q", s.UserAgent, "Mozilla/5.0")
		}
	})

	t.Run("accepts empty IP and User-Agent values", func(t *testing.T) {
		s, _, err := NewSession("1", "user-1", "", "", time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.IPAddress != "" || s.UserAgent != "" {
			t.Error("Empty IPAddress/UserAgent values should not cause a validation error")
		}
	})

	t.Run("rejects an empty userID", func(t *testing.T) {
		_, _, err := NewSession("1", "", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if err != ErrEmptyUserID {
			t.Errorf("error = %v, want %v", err, ErrEmptyUserID)
		}
	})

	t.Run("rejects a zero or negative TTL", func(t *testing.T) {
		_, _, err := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", 0)
		if err != ErrInvalidTTL {
			t.Errorf("error = %v, want %v", err, ErrInvalidTTL)
		}

		_, _, err = NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", -time.Hour)
		if err != ErrInvalidTTL {
			t.Errorf("error = %v, want %v", err, ErrInvalidTTL)
		}
	})

	t.Run("generates a different token on each call", func(t *testing.T) {
		_, token1, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		_, token2, _ := NewSession("2", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if token1 == token2 {
			t.Error("two sessions should never receive the same token")
		}
	})
}

func TestSession_VerifyToken(t *testing.T) {
	s, plainToken, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)

	if !s.VerifyToken(plainToken) {
		t.Error("VerifyToken should return true for the correct token")
	}
	if s.VerifyToken("incorrect-token") {
		t.Error("VerifyToken should return false for an incorrect token")
	}
}

func TestSession_IsExpired(t *testing.T) {
	t.Run("a fresh session is not expired", func(t *testing.T) {
		s, _, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if s.IsExpired() {
			t.Error("a session valid for 1h should not be expired")
		}
	})

	t.Run("a session with ExpiresAt in the past is expired", func(t *testing.T) {
		s, _, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		if !s.IsExpired() {
			t.Error("a session whose ExpiresAt is in the past should be expired")
		}
	})
}

func TestSession_Touch(t *testing.T) {
	s, _, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
	original := s.LastUsedAt
	originalExpiry := s.ExpiresAt

	time.Sleep(time.Millisecond)
	s.Touch()

	if !s.LastUsedAt.After(original) {
		t.Error("LastUsedAt should have advanced after Touch")
	}
	if s.ExpiresAt != originalExpiry {
		t.Error("Touch should NOT modify ExpiresAt (no sliding session)")
	}
}

func TestSession_IsSuspicious(t *testing.T) {
	const originalIP = "203.0.113.1"
	const originalUA = "Mozilla/5.0 (Windows)"

	tests := []struct {
		name      string
		ip        string
		userAgent string
		want      bool
	}{
		{"identical IP and User-Agent", originalIP, originalUA, false},
		{"only the IP changes (mobile network)", "198.51.100.9", originalUA, false},
		{"only the User-Agent changes (browser update)", originalIP, "Mozilla/5.0 (Mac)", false},
		{"IP et User-Agent changent tous les deux", "198.51.100.9", "Mozilla/5.0 (Mac)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, _ := NewSession("1", "user-1", originalIP, originalUA, time.Hour)

			got := s.IsSuspicious(tt.ip, tt.userAgent)
			if got != tt.want {
				t.Errorf("IsSuspicious(%q, %q) = %v, want %v", tt.ip, tt.userAgent, got, tt.want)
			}
		})
	}
}
