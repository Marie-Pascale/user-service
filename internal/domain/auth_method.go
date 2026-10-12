package domain

import (
	"errors"
	"time"
)

type Provider string

const (
	ProviderLocal    Provider = "local"
	ProviderGoogle   Provider = "google"
	ProviderFacebook Provider = "facebook"
)

var (
	ErrEmptyUserID           = errors.New("userID is required")
	ErrEmptyPassword         = errors.New("password is required")
	ErrEmptyProviderUserID   = errors.New("providerUserID is required for third-party providers")
	ErrInvalidProvider       = errors.New("invalid provider")
	ErrPasswordNotApplicable = errors.New("password is not applicable for third-party providers")
)

type AuthMethod struct {
	ID             string
	UserID         string
	Provider       Provider
	PasswordHash   string
	ProviderUserID string
	CreatedAt      time.Time
	LastUsedAt     time.Time
}

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

func (a *AuthMethod) Authenticate(password string, hasher PasswordHasher) bool {
	if a.Provider != ProviderLocal {
		return false
	}
	return hasher.Verify(password, a.PasswordHash)
}

func (a *AuthMethod) MarkUsed() {
	a.LastUsedAt = time.Now().UTC()
}
