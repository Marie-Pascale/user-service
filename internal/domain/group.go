package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrEmptyGroupName = errors.New("group name is required")

type Group struct {
	ID            string
	Name          string
	Description   string
	MemberUserIDs []string
	RoleIDs       []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

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

func (g *Group) AddMember(userID string) {
	if g.HasMember(userID) {
		return
	}
	g.MemberUserIDs = append(g.MemberUserIDs, userID)
	g.UpdatedAt = time.Now().UTC()
}

func (g *Group) RemoveMember(userID string) {
	for i, id := range g.MemberUserIDs {
		if id == userID {
			g.MemberUserIDs = append(g.MemberUserIDs[:i], g.MemberUserIDs[i+1:]...)
			g.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

func (g *Group) HasMember(userID string) bool {
	for _, id := range g.MemberUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func (g *Group) AssignRole(roleID string) {
	for _, id := range g.RoleIDs {
		if id == roleID {
			return
		}
	}
	g.RoleIDs = append(g.RoleIDs, roleID)
	g.UpdatedAt = time.Now().UTC()
}

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
