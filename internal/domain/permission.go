package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyServiceID     = errors.New("serviceID is required")
	ErrEmptyPermissionKey = errors.New("permission key is required")
)

type Permission struct {
	ID          string
	ServiceID   string
	Key         string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewPermission(id, serviceID, key, description string) (*Permission, error) {
	if serviceID == "" {
		return nil, ErrEmptyServiceID
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrEmptyPermissionKey
	}

	now := time.Now().UTC()
	return &Permission{
		ID:          id,
		ServiceID:   serviceID,
		Key:         key,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
