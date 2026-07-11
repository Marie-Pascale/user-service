package domain

import (
	"errors"
	"time"
)

var ErrInvalidTTL = errors.New("session TTL must be positive")

// Session represente une connexion active d'un usager. Contrairement a un
// JWT (signe, sans etat, non revocable avant expiration), une Session est
// un jeton OPAQUE stocke en base -- choix fait volontairement pour pouvoir
// revoquer une session immediatement (deconnexion, compromission suspectee)
// sans attendre son expiration naturelle.
//
// Comme pour Service.APIKeyHash, le jeton en clair n'est JAMAIS stocke --
// seul son hash SHA-256 (voir token.go, partage avec Service).
type Session struct {
	ID         string
	UserID     string
	TokenHash  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt time.Time
}

// NewSession cree une nouvelle session pour un usager, valide pour la
// duree ttl. Le jeton en clair (plainToken) n'est retourne qu'UNE SEULE
// FOIS, a la creation -- exactement comme domain.NewService pour les cles
// API. C'est ce jeton que l'usager doit presenter (via l'entete
// Authorization) pour prouver qu'il est bien connecte.
func NewSession(id, userID string, ttl time.Duration) (session *Session, plainToken string, err error) {
	if userID == "" {
		return nil, "", ErrEmptyUserID
	}
	if ttl <= 0 {
		return nil, "", ErrInvalidTTL
	}

	plainToken, err = generateSecureToken(32)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	session = &Session{
		ID:         id,
		UserID:     userID,
		TokenHash:  HashToken(plainToken),
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		LastUsedAt: now,
	}
	return session, plainToken, nil
}

// VerifyToken retourne true si le jeton en clair fourni correspond a
// cette session.
func (s *Session) VerifyToken(plainToken string) bool {
	return HashToken(plainToken) == s.TokenHash
}

// IsExpired indique si la session a depasse sa duree de vie.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// Touch met a jour l'horodatage de derniere utilisation. N'etend PAS
// ExpiresAt (pas de "sliding session" pour l'instant) -- une session
// expire toujours a duree fixe depuis sa creation, ce qui reste simple a
// raisonner. Ca pourra evoluer si le besoin se presente.
func (s *Session) Touch() {
	s.LastUsedAt = time.Now().UTC()
}
