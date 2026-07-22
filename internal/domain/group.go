package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrEmptyGroupName = errors.New("group name is required")

// Group est un regroupement d'usagers. Les roles assignes a un groupe
// (RoleIDs) sont herites par tous ses membres -- c'est le principal
// interet des groupes : assigner un role a 50 usagers d'un coup plutot
// que 50 fois individuellement.
type Group struct {
	ID            string
	Name          string
	Description   string
	MemberUserIDs []string
	RoleIDs       []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewGroup cree un groupe valide, vide au depart.
func NewGroup(id, name, description string) (*Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyGroupName
	}

	now := time.Now().UTC()
	return &Group{
		ID:            id,
		Name:          name,
		Description:   description,
		MemberUserIDs: []string{},
		RoleIDs:       []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// AddMember ajoute un usager au groupe, s'il n'en est pas deja membre.
func (g *Group) AddMember(userID string) {
	if g.HasMember(userID) {
		return
	}
	g.MemberUserIDs = append(g.MemberUserIDs, userID)
	g.UpdatedAt = time.Now().UTC()
}

// RemoveMember retire un usager du groupe.
func (g *Group) RemoveMember(userID string) {
	for i, id := range g.MemberUserIDs {
		if id == userID {
			g.MemberUserIDs = append(g.MemberUserIDs[:i], g.MemberUserIDs[i+1:]...)
			g.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// HasMember indique si un usager fait partie du groupe.
func (g *Group) HasMember(userID string) bool {
	for _, id := range g.MemberUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// AssignRole assigne un role au groupe (herite par tous ses membres).
func (g *Group) AssignRole(roleID string) {
	for _, id := range g.RoleIDs {
		if id == roleID {
			return
		}
	}
	g.RoleIDs = append(g.RoleIDs, roleID)
	g.UpdatedAt = time.Now().UTC()
}

// UpdateDetails modifie le nom et la description d'un groupe existant,
// avec la meme validation que NewGroup. Ne touche jamais MemberUserIDs ni
// RoleIDs -- geres separement via AddMember/RemoveMember/AssignRole.
func (g *Group) UpdateDetails(name, description string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyGroupName
	}
	g.Name = name
	g.Description = description
	g.UpdatedAt = time.Now().UTC()
	return nil
}
