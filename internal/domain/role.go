package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrEmptyRoleName = errors.New("role name is required")

type Role struct {
	ID            string
	Name          string
	Description   string
	PermissionIDs []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewRole(id, name, description string) (*Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyRoleName
	}

	now := time.Now().UTC()
	return &Role{
		ID:            id,
		Name:          name,
		Description:   description,
		PermissionIDs: []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (r *Role) GrantPermission(permissionID string) {
	for _, id := range r.PermissionIDs {
		if id == permissionID {
			return
		}
	}
	r.PermissionIDs = append(r.PermissionIDs, permissionID)
	r.UpdatedAt = time.Now().UTC()
}

func (r *Role) RevokePermission(permissionID string) {
	for i, id := range r.PermissionIDs {
		if id == permissionID {
			r.PermissionIDs = append(r.PermissionIDs[:i], r.PermissionIDs[i+1:]...)
			r.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

func (r *Role) HasPermission(permissionID string) bool {
	for _, id := range r.PermissionIDs {
		if id == permissionID {
			return true
		}
	}
	return false
}

func (r *Role) UpdateDetails(name, description string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyRoleName
	}
	r.Name = name
	r.Description = description
	r.UpdatedAt = time.Now().UTC()
	return nil
}
