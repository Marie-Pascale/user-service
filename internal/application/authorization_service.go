package application

import (
	"context"
)

type AuthorizationService struct {
	Users       UserRepository
	Roles       RoleRepository
	Groups      GroupRepository
	Permissions PermissionRepository
}

func NewAuthorizationService(users UserRepository, roles RoleRepository, groups GroupRepository, permissions PermissionRepository) *AuthorizationService {
	return &AuthorizationService{Users: users, Roles: roles, Groups: groups, Permissions: permissions}
}

func (s *AuthorizationService) ResolvePermissionKeys(ctx context.Context, userID, serviceID string) ([]string, error) {
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleIDs := newStringSet(u.RoleIDs...)

	groups, err := s.Groups.ListByMemberUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		roleIDs.addAll(g.RoleIDs)
	}

	permissionIDs := newStringSet(u.PermissionIDs...)

	for id := range roleIDs {
		role, err := s.Roles.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		permissionIDs.addAll(role.PermissionIDs)
	}

	permissions, err := s.Permissions.ListByIDs(ctx, permissionIDs.slice())
	if err != nil {
		return nil, err
	}

	keys := []string{}
	for _, p := range permissions {
		if p.ServiceID == serviceID {
			keys = append(keys, p.Key)
		}
	}
	return keys, nil
}

func (s *AuthorizationService) HasPermission(ctx context.Context, userID, serviceID, permissionKey string) (bool, error) {
	keys, err := s.ResolvePermissionKeys(ctx, userID, serviceID)
	if err != nil {
		return false, err
	}
	for _, k := range keys {
		if k == permissionKey {
			return true, nil
		}
	}
	return false, nil
}

type stringSet map[string]struct{}

func newStringSet(values ...string) stringSet {
	s := make(stringSet, len(values))
	s.addAll(values)
	return s
}

func (s stringSet) addAll(values []string) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

func (s stringSet) slice() []string {
	out := make([]string, 0, len(s))
	for v := range s {
		out = append(out, v)
	}
	return out
}
