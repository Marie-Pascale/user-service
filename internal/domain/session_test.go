package domain

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	t.Run("cree une session valide", func(t *testing.T) {
		s, plainToken, err := NewSession("1", "user-1", time.Hour)
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
	})

	t.Run("rejette un userID vide", func(t *testing.T) {
		_, _, err := NewSession("1", "", time.Hour)
		if err != ErrEmptyUserID {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyUserID)
		}
	})

	t.Run("rejette une duree de vie nulle ou negative", func(t *testing.T) {
		_, _, err := NewSession("1", "user-1", 0)
		if err != ErrInvalidTTL {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidTTL)
		}

		_, _, err = NewSession("1", "user-1", -time.Hour)
		if err != ErrInvalidTTL {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidTTL)
		}
	})

	t.Run("genere un jeton different a chaque appel", func(t *testing.T) {
		_, token1, _ := NewSession("1", "user-1", time.Hour)
		_, token2, _ := NewSession("2", "user-1", time.Hour)
		if token1 == token2 {
			t.Error("deux sessions ne devraient jamais recevoir le meme jeton")
		}
	})
}

func TestSession_VerifyToken(t *testing.T) {
	s, plainToken, _ := NewSession("1", "user-1", time.Hour)

	if !s.VerifyToken(plainToken) {
		t.Error("VerifyToken devrait retourner true pour le bon jeton")
	}
	if s.VerifyToken("mauvais-jeton") {
		t.Error("VerifyToken devrait retourner false pour un mauvais jeton")
	}
}

func TestSession_IsExpired(t *testing.T) {
	t.Run("une session fraiche n'est pas expiree", func(t *testing.T) {
		s, _, _ := NewSession("1", "user-1", time.Hour)
		if s.IsExpired() {
			t.Error("une session valide pour 1h ne devrait pas etre expiree")
		}
	})

	t.Run("une session avec ExpiresAt dans le passe est expiree", func(t *testing.T) {
		s, _, _ := NewSession("1", "user-1", time.Hour)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		if !s.IsExpired() {
			t.Error("une session dont ExpiresAt est dans le passe devrait etre expiree")
		}
	})
}

func TestSession_Touch(t *testing.T) {
	s, _, _ := NewSession("1", "user-1", time.Hour)
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
