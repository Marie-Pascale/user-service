package domain

import (
	"errors"
	"time"
)

var ErrEmptyServiceName = errors.New("service name is required")

type Service struct {
	ID          string
	Name        string
	Description string
	APIKeyHash  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewService(id, name, description string) (svc *Service, plainAPIKey string, err error) {
	if name == "" {
		return nil, "", ErrEmptyServiceName
	}

	plainAPIKey, err = generateSecureToken(32)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	svc = &Service{
		ID:          id,
		Name:        name,
		Description: description,
		APIKeyHash:  HashToken(plainAPIKey),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return svc, plainAPIKey, nil
}

func (s *Service) VerifyAPIKey(plainAPIKey string) bool {
	return HashToken(plainAPIKey) == s.APIKeyHash
}

func (s *Service) RotateAPIKey() (plainAPIKey string, err error) {
	plainAPIKey, err = generateSecureToken(32)
	if err != nil {
		return "", err
	}
	s.APIKeyHash = HashToken(plainAPIKey)
	s.UpdatedAt = time.Now().UTC()
	return plainAPIKey, nil
}

func (s *Service) UpdateDetails(name, description string) error {
	if name == "" {
		return ErrEmptyServiceName
	}
	s.Name = name
	s.Description = description
	s.UpdatedAt = time.Now().UTC()
	return nil
}
