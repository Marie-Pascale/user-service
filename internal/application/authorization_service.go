package application

import (
	"context"
)

// AuthorizationService resout les permissions EFFECTIVES d'un usager pour
// un service donne : la combinaison de ses permissions directes, de celles
// heritees de ses roles directs, et de celles heritees des roles de ses
// groupes -- exactement la regle qu'on avait definie en concevant le
// domaine (voir User.PermissionIDs, User.RoleIDs, Group.RoleIDs).
//
// C'est ICI, dans la couche application, que cette regle est calculee --
// pas dans le domaine, parce qu'elle a besoin d'aller chercher des
// entites liees (les Role et Group complets) via plusieurs repositories.
// Le domaine ne connait que l'objet qu'on lui donne ; l'orchestration
// entre plusieurs entites est le role de l'application.
type AuthorizationService struct {
	Users       UserRepository
	Roles       RoleRepository
	Groups      GroupRepository
	Permissions PermissionRepository
}

func NewAuthorizationService(users UserRepository, roles RoleRepository, groups GroupRepository, permissions PermissionRepository) *AuthorizationService {
	return &AuthorizationService{Users: users, Roles: roles, Groups: groups, Permissions: permissions}
}

// ResolvePermissionKeys retourne les cles de permission (ex: "orders:read")
// que l'usager possede effectivement pour un service donne.
func (s *AuthorizationService) ResolvePermissionKeys(ctx context.Context, userID, serviceID string) ([]string, error) {
	u, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Etape 1 : rassembler tous les roles applicables -- directs + herites
	// des groupes dont l'usager est membre.
	roleIDs := newStringSet(u.RoleIDs...)

	groups, err := s.Groups.ListByMemberUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		roleIDs.addAll(g.RoleIDs)
	}

	// Etape 2 : rassembler toutes les permissions -- directes + celles de
	// chaque role rassemble a l'etape 1.
	permissionIDs := newStringSet(u.PermissionIDs...)

	for id := range roleIDs {
		role, err := s.Roles.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		permissionIDs.addAll(role.PermissionIDs)
	}

	// Etape 3 : charger les permissions completes et filtrer par service --
	// une permission ID rassemblee peut appartenir a n'importe quel
	// service, on ne veut que celles du service demande.
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

// HasPermission est un raccourci pratique pour le cas le plus courant :
// verifier qu'un usager a UNE permission precise pour un service donne
// (ex: dans un middleware d'autorisation).
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

// stringSet est un petit utilitaire local pour dedupliquer des IDs sans
// repeter la meme boucle "si pas deja present, ajouter" partout ci-dessus.
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
