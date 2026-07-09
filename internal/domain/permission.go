package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyServiceID     = errors.New("serviceID is required")
	ErrEmptyPermissionKey = errors.New("permission key is required")
)

// Permission represente une action precise, definie par un service.
// La cle suit generalement la convention "ressource:action"
// (ex: "orders:read", "orders:write", "billing:refund").
//
// Une permission n'a de sens que dans le contexte d'un service donne --
// deux services peuvent tres bien definir chacun une permission "read"
// sans qu'il y ait de conflit, puisqu'elles sont scopees par ServiceID.
type Permission struct {
	ID          string
	ServiceID   string
	Key         string
	Description string
}

// NewPermission cree une permission valide.
func NewPermission(id, serviceID, key, description string) (*Permission, error) {
	if serviceID == "" {
		return nil, ErrEmptyServiceID
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrEmptyPermissionKey
	}

	return &Permission{
		ID:          id,
		ServiceID:   serviceID,
		Key:         key,
		Description: description,
	}, nil
}
