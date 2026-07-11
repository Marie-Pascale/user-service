package domain

import (
	"errors"
	"time"
)

var ErrEmptyServiceName = errors.New("service name is required")

// Service represente un service tiers de l'ecosysteme qui consomme
// l'API des permissions (ex: "orders-service", "billing-service").
// Chaque service est identifie par son propre ID et s'authentifie
// aupres de nous via une cle API (voir token.go pour la generation et
// le hachage, partages avec Session).
type Service struct {
	ID          string
	Name        string
	Description string
	APIKeyHash  string
	CreatedAt   time.Time
}

// NewService cree un nouveau service et genere sa cle API.
// La cle en clair (plainAPIKey) n'est retournee qu'UNE SEULE FOIS, a la
// creation -- elle n'est jamais recuperable par la suite. C'est a
// l'appelant de la communiquer de maniere securisee au service concerne.
func NewService(id, name, description string) (svc *Service, plainAPIKey string, err error) {
	if name == "" {
		return nil, "", ErrEmptyServiceName
	}

	plainAPIKey, err = generateSecureToken(32)
	if err != nil {
		return nil, "", err
	}

	svc = &Service{
		ID:          id,
		Name:        name,
		Description: description,
		APIKeyHash:  HashToken(plainAPIKey),
		CreatedAt:   time.Now().UTC(),
	}
	return svc, plainAPIKey, nil
}

// VerifyAPIKey retourne true si la cle en clair fournie correspond
// a celle de ce service.
func (s *Service) VerifyAPIKey(plainAPIKey string) bool {
	return HashToken(plainAPIKey) == s.APIKeyHash
}

// RotateAPIKey genere une nouvelle cle API pour ce service, invalidant
// l'ancienne. Retourne la nouvelle cle en clair (une seule fois, comme
// a la creation).
func (s *Service) RotateAPIKey() (plainAPIKey string, err error) {
	plainAPIKey, err = generateSecureToken(32)
	if err != nil {
		return "", err
	}
	s.APIKeyHash = HashToken(plainAPIKey)
	return plainAPIKey, nil
}
