package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyName    = errors.New("name is required")
	ErrInvalidEmail = errors.New("invalid email")
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

type User struct {
	ID            string
	Name          string
	Email         string
	Status        Status
	RoleIDs       []string
	PermissionIDs []string
	CreatedAt     time.Time
	UpdatedAt     time.Time

	LastLoginAt *time.Time
}

func NewUser(id, name, email string) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	now := time.Now().UTC()
	return &User{
		ID:            id,
		Name:          name,
		Email:         email,
		Status:        StatusActive,
		RoleIDs:       []string{},
		PermissionIDs: []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	domain := email[at+1:]
	return strings.Contains(domain, ".")
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) AssignRole(roleID string) {
	for _, id := range u.RoleIDs {
		if id == roleID {
			return
		}
	}
	u.RoleIDs = append(u.RoleIDs, roleID)
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) AssignPermission(permissionID string) {
	for _, id := range u.PermissionIDs {
		if id == permissionID {
			return
		}
	}
	u.PermissionIDs = append(u.PermissionIDs, permissionID)
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) UpdateProfile(name, email string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	u.Name = name
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}
