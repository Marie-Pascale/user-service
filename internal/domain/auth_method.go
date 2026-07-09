package domain

import (
	"errors"
	"time"
)

// Provider identifie la methode utilisee pour s'authentifier.
type Provider string

const (
	ProviderLocal    Provider = "local"
	ProviderGoogle   Provider = "google"
	ProviderFacebook Provider = "facebook"
)

var (
	ErrEmptyUserID          = errors.New("userID is required")
	ErrEmptyPassword        = errors.New("password is required")
	ErrEmptyProviderUserID  = errors.New("providerUserID is required for third-party providers")
	ErrInvalidProvider      = errors.New("invalid provider")
	ErrPasswordNotApplicable = errors.New("password is not applicable for third-party providers")
)

// AuthMethod represente UNE facon pour un usager de s'authentifier.
// Un meme User peut avoir plusieurs AuthMethod (ex: mot de passe local
// ET connexion Google) -- c'est pourquoi cette entite est separee de
// User plutot que d'etre des champs directement dessus.
type AuthMethod struct {
	ID             string
	UserID         string
	Provider       Provider
	PasswordHash   string // rempli seulement si Provider == ProviderLocal
	ProviderUserID string // rempli seulement si Provider != ProviderLocal (l'ID chez le tiers)
	CreatedAt      time.Time
	LastUsedAt     time.Time
}

// NewLocalAuthMethod cree une methode d'authentification par mot de passe.
// Le mot de passe en clair n'est jamais stocke : il est immediatement
// hache via le PasswordHasher fourni (implemente dans l'infrastructure,
// ex: bcrypt).
func NewLocalAuthMethod(id, userID, password string, hasher PasswordHasher) (*AuthMethod, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if password == "" {
		return nil, ErrEmptyPassword
	}

	hash, err := hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &AuthMethod{
		ID:           id,
		UserID:       userID,
		Provider:     ProviderLocal,
		PasswordHash: hash,
		CreatedAt:    now,
		LastUsedAt:   now,
	}, nil
}

// NewOAuthMethod cree une methode d'authentification via un fournisseur
// tiers (Google, Facebook...). providerUserID est l'identifiant unique
// de l'usager chez ce fournisseur (ex: le "sub" d'un token Google).
func NewOAuthMethod(id, userID string, provider Provider, providerUserID string) (*AuthMethod, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if provider != ProviderGoogle && provider != ProviderFacebook {
		return nil, ErrInvalidProvider
	}
	if providerUserID == "" {
		return nil, ErrEmptyProviderUserID
	}

	now := time.Now().UTC()
	return &AuthMethod{
		ID:             id,
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		CreatedAt:      now,
		LastUsedAt:     now,
	}, nil
}

// Authenticate verifie un mot de passe en clair contre le hash stocke.
// Ne s'applique qu'aux methodes locales -- retourne toujours false pour
// les methodes tierces (l'authentification tierce se fait par un autre
// mecanisme, ex: verification de token OAuth, pas par mot de passe).
func (a *AuthMethod) Authenticate(password string, hasher PasswordHasher) bool {
	if a.Provider != ProviderLocal {
		return false
	}
	return hasher.Verify(password, a.PasswordHash)
}

// MarkUsed met a jour l'horodatage de derniere utilisation.
func (a *AuthMethod) MarkUsed() {
	a.LastUsedAt = time.Now().UTC()
}
