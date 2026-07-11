package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var ErrEmptyServiceName = errors.New("service name is required")

// Service represente un service tiers de l'ecosysteme qui consomme
// l'API des permissions (ex: "orders-service", "billing-service").
// Chaque service est identifie par son propre ID et s'authentifie
// aupres de nous via une cle API.
//
// La cle API n'est JAMAIS stockee en clair -- seul son hash SHA-256
// est conserve. On utilise SHA-256 plutot que bcrypt ici parce qu'une
// cle API generee aleatoirement a deja une entropie tres elevee (contrairement
// a un mot de passe choisi par un humain) : un hash lent concu pour
// resister au brute-force n'apporte rien de plus, et couterait du
// temps CPU inutile a chaque requete entrante.
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

	plainAPIKey, err = generateAPIKey()
	if err != nil {
		return nil, "", err
	}

	svc = &Service{
		ID:          id,
		Name:        name,
		Description: description,
		APIKeyHash:  HashAPIKey(plainAPIKey),
		CreatedAt:   time.Now().UTC(),
	}
	return svc, plainAPIKey, nil
}

// VerifyAPIKey retourne true si la cle en clair fournie correspond
// a celle de ce service.
func (s *Service) VerifyAPIKey(plainAPIKey string) bool {
	return HashAPIKey(plainAPIKey) == s.APIKeyHash
}

// RotateAPIKey genere une nouvelle cle API pour ce service, invalidant
// l'ancienne. Retourne la nouvelle cle en clair (une seule fois, comme
// a la creation).
func (s *Service) RotateAPIKey() (plainAPIKey string, err error) {
	plainAPIKey, err = generateAPIKey()
	if err != nil {
		return "", err
	}
	s.APIKeyHash = HashAPIKey(plainAPIKey)
	return plainAPIKey, nil
}

// generateAPIKey produit 32 octets aleatoires cryptographiquement surs,
// encodes en hexadecimal (64 caracteres).
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashAPIKey hache une cle API en clair avec SHA-256. Exportee (pas
// hashAPIKey) parce que le middleware d'authentification des services,
// dans la couche infrastructure, doit hacher la cle recue AVANT de la
// chercher en base via ServiceRepository.GetByAPIKeyHash -- la recherche
// se fait par hash, jamais par cle en clair.
func HashAPIKey(plainAPIKey string) string {
	sum := sha256.Sum256([]byte(plainAPIKey))
	return hex.EncodeToString(sum[:])
}
