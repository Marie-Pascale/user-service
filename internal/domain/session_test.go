package domain

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	t.Run("cree une session valide", func(t *testing.T) {
		s, plainToken, err := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if plainToken == "" {
			t.Error("le jeton en clair ne devrait pas etre vide")
		}
		if s.TokenHash == plainToken {
			t.Error("TokenHash ne devrait jamais etre egal au jeton en clair")
		}
		if !s.ExpiresAt.After(s.CreatedAt) {
			t.Error("ExpiresAt devrait etre apres CreatedAt")
		}
		if s.IPAddress != "203.0.113.1" {
			t.Errorf("IPAddress = %q, attendu %q", s.IPAddress, "203.0.113.1")
		}
		if s.UserAgent != "Mozilla/5.0" {
			t.Errorf("UserAgent = %q, attendu %q", s.UserAgent, "Mozilla/5.0")
		}
	})

	t.Run("accepte une IP et un User-Agent vides", func(t *testing.T) {
		s, _, err := NewSession("1", "user-1", "", "", time.Hour)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if s.IPAddress != "" || s.UserAgent != "" {
			t.Error("IPAddress/UserAgent vides ne devraient pas etre une erreur de validation")
		}
	})

	t.Run("rejette un userID vide", func(t *testing.T) {
		_, _, err := NewSession("1", "", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if err != ErrEmptyUserID {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyUserID)
		}
	})

	t.Run("rejette une duree de vie nulle ou negative", func(t *testing.T) {
		_, _, err := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", 0)
		if err != ErrInvalidTTL {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidTTL)
		}

		_, _, err = NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", -time.Hour)
		if err != ErrInvalidTTL {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidTTL)
		}
	})

	t.Run("genere un jeton different a chaque appel", func(t *testing.T) {
		_, token1, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		_, token2, _ := NewSession("2", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if token1 == token2 {
			t.Error("deux sessions ne devraient jamais recevoir le meme jeton")
		}
	})
}

func TestSession_VerifyToken(t *testing.T) {
	s, plainToken, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)

	if !s.VerifyToken(plainToken) {
		t.Error("VerifyToken devrait retourner true pour le bon jeton")
	}
	if s.VerifyToken("mauvais-jeton") {
		t.Error("VerifyToken devrait retourner false pour un mauvais jeton")
	}
}

func TestSession_IsExpired(t *testing.T) {
	t.Run("une session fraiche n'est pas expiree", func(t *testing.T) {
		s, _, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		if s.IsExpired() {
			t.Error("une session valide pour 1h ne devrait pas etre expiree")
		}
	})

	t.Run("une session avec ExpiresAt dans le passe est expiree", func(t *testing.T) {
		s, _, _ := NewSession("1", "user-1", "203.0.113.1", "Mozilla/5.0", time.Hour)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		if !s.IsExpired() {
			t.Error("une session dont ExpiresAt est dans le passe devrait etre expiree")
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
		t.Error("LastUsedAt devrait avoir avance apres Touch")
	}
	if s.ExpiresAt != originalExpiry {
		t.Error("Touch ne devrait PAS modifier ExpiresAt (pas de sliding session)")
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
		{"IP et User-Agent identiques", originalIP, originalUA, false},
		{"seule l'IP change (reseau mobile)", "198.51.100.9", originalUA, false},
		{"seul le User-Agent change (mise a jour navigateur)", originalIP, "Mozilla/5.0 (Mac)", false},
		{"IP et User-Agent changent tous les deux", "198.51.100.9", "Mozilla/5.0 (Mac)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, _ := NewSession("1", "user-1", originalIP, originalUA, time.Hour)

			got := s.IsSuspicious(tt.ip, tt.userAgent)
			if got != tt.want {
				t.Errorf("IsSuspicious(%q, %q) = %v, attendu %v", tt.ip, tt.userAgent, got, tt.want)
			}
		})
	}
}
