package httpapi

import (
	"context"

	"user-service/internal/application"
	"user-service/internal/domain"
)

type fakeUserRepo struct {
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[string]*domain.User{}, byEmail: map[string]*domain.User{}}
}

func (f *fakeUserRepo) Create(ctx context.Context, u *domain.User) error {
	f.byID[u.ID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return nil, application.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) List(ctx context.Context) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(f.byID))
	for _, u := range f.byID {
		users = append(users, u)
	}
	return users, nil
}

func (f *fakeUserRepo) Update(ctx context.Context, u *domain.User) error {
	if _, ok := f.byID[u.ID]; !ok {
		return application.ErrNotFound
	}
	f.byID[u.ID] = u
	f.byEmail[u.Email] = u
	return nil
}

func (f *fakeUserRepo) Delete(ctx context.Context, id string) error {
	u, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	delete(f.byEmail, u.Email)
	return nil
}

func (f *fakeUserRepo) AssignRole(ctx context.Context, userID, roleID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return application.ErrNotFound
	}
	u.AssignRole(roleID)
	return nil
}

func (f *fakeUserRepo) RevokeRole(ctx context.Context, userID, roleID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return application.ErrNotFound
	}
	for i, id := range u.RoleIDs {
		if id == roleID {
			u.RoleIDs = append(u.RoleIDs[:i], u.RoleIDs[i+1:]...)
			break
		}
	}
	return nil
}

func (f *fakeUserRepo) AssignPermission(ctx context.Context, userID, permissionID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return application.ErrNotFound
	}
	u.AssignPermission(permissionID)
	return nil
}

func (f *fakeUserRepo) RevokePermission(ctx context.Context, userID, permissionID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return application.ErrNotFound
	}
	for i, id := range u.PermissionIDs {
		if id == permissionID {
			u.PermissionIDs = append(u.PermissionIDs[:i], u.PermissionIDs[i+1:]...)
			break
		}
	}
	return nil
}

func (f *fakeUserRepo) RecordLogin(ctx context.Context, userID string) error {
	u, ok := f.byID[userID]
	if !ok {
		return application.ErrNotFound
	}
	u.RecordLogin()
	return nil
}

var _ application.UserRepository = (*fakeUserRepo)(nil)

type fakeAuthMethodRepo struct {
	byID   map[string]*domain.AuthMethod
	byUser map[string][]*domain.AuthMethod
}

func newFakeAuthMethodRepo() *fakeAuthMethodRepo {
	return &fakeAuthMethodRepo{byID: map[string]*domain.AuthMethod{}, byUser: map[string][]*domain.AuthMethod{}}
}

func (f *fakeAuthMethodRepo) Create(ctx context.Context, am *domain.AuthMethod) error {
	f.byID[am.ID] = am
	f.byUser[am.UserID] = append(f.byUser[am.UserID], am)
	return nil
}

func (f *fakeAuthMethodRepo) GetByID(ctx context.Context, id string) (*domain.AuthMethod, error) {
	am, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return am, nil
}

func (f *fakeAuthMethodRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.AuthMethod, error) {
	return f.byUser[userID], nil
}

func (f *fakeAuthMethodRepo) GetByProviderUserID(ctx context.Context, provider domain.Provider, providerUserID string) (*domain.AuthMethod, error) {
	for _, am := range f.byID {
		if am.Provider == provider && am.ProviderUserID == providerUserID {
			return am, nil
		}
	}
	return nil, application.ErrNotFound
}

func (f *fakeAuthMethodRepo) MarkUsed(ctx context.Context, id string) error {
	if _, ok := f.byID[id]; !ok {
		return application.ErrNotFound
	}
	return nil
}

func (f *fakeAuthMethodRepo) Delete(ctx context.Context, id string) error {
	am, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	list := f.byUser[am.UserID]
	for i, x := range list {
		if x.ID == id {
			f.byUser[am.UserID] = append(list[:i], list[i+1:]...)
			break
		}
	}
	return nil
}

var _ application.AuthMethodRepository = (*fakeAuthMethodRepo)(nil)

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakeHasher) Verify(password, hash string) bool {
	return "hashed:"+password == hash
}

type fakeServiceRepo struct {
	byID      map[string]*domain.Service
	byKeyHash map[string]*domain.Service
}

func newFakeServiceRepo() *fakeServiceRepo {
	return &fakeServiceRepo{byID: map[string]*domain.Service{}, byKeyHash: map[string]*domain.Service{}}
}

func (f *fakeServiceRepo) Create(ctx context.Context, s *domain.Service) error {
	f.byID[s.ID] = s
	f.byKeyHash[s.APIKeyHash] = s
	return nil
}

func (f *fakeServiceRepo) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	s, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return s, nil
}

func (f *fakeServiceRepo) GetByAPIKeyHash(ctx context.Context, hash string) (*domain.Service, error) {
	s, ok := f.byKeyHash[hash]
	if !ok {
		return nil, application.ErrNotFound
	}
	return s, nil
}

func (f *fakeServiceRepo) List(ctx context.Context) ([]*domain.Service, error) {
	out := make([]*domain.Service, 0, len(f.byID))
	for _, s := range f.byID {
		out = append(out, s)
	}
	return out, nil
}

func (f *fakeServiceRepo) Update(ctx context.Context, s *domain.Service) error {
	if _, ok := f.byID[s.ID]; !ok {
		return application.ErrNotFound
	}
	f.byID[s.ID] = s
	return nil
}

func (f *fakeServiceRepo) UpdateAPIKeyHash(ctx context.Context, id, newHash string) error {
	s, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	for hash, svc := range f.byKeyHash {
		if svc.ID == id {
			delete(f.byKeyHash, hash)
		}
	}
	s.APIKeyHash = newHash
	f.byKeyHash[newHash] = s
	return nil
}

func (f *fakeServiceRepo) Delete(ctx context.Context, id string) error {
	s, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	delete(f.byKeyHash, s.APIKeyHash)
	return nil
}

var _ application.ServiceRepository = (*fakeServiceRepo)(nil)

type fakeSessionRepo struct {
	byID        map[string]*domain.Session
	byTokenHash map[string]*domain.Session
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{byID: map[string]*domain.Session{}, byTokenHash: map[string]*domain.Session{}}
}

func (f *fakeSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	f.byID[s.ID] = s
	f.byTokenHash[s.TokenHash] = s
	return nil
}

func (f *fakeSessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	s, ok := f.byTokenHash[tokenHash]
	if !ok {
		return nil, application.ErrNotFound
	}
	return s, nil
}

func (f *fakeSessionRepo) Touch(ctx context.Context, id string) error {
	s, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	s.Touch()
	return nil
}

func (f *fakeSessionRepo) Delete(ctx context.Context, id string) error {
	s, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	delete(f.byTokenHash, s.TokenHash)
	return nil
}

func (f *fakeSessionRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	for id, s := range f.byID {
		if s.UserID == userID {
			delete(f.byID, id)
			delete(f.byTokenHash, s.TokenHash)
		}
	}
	return nil
}

var _ application.SessionRepository = (*fakeSessionRepo)(nil)

type fakeRoleRepo struct {
	byID map[string]*domain.Role
}

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{byID: map[string]*domain.Role{}}
}

func (f *fakeRoleRepo) Create(ctx context.Context, r *domain.Role) error {
	f.byID[r.ID] = r
	return nil
}

func (f *fakeRoleRepo) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return r, nil
}

func (f *fakeRoleRepo) List(ctx context.Context) ([]*domain.Role, error) {
	out := make([]*domain.Role, 0, len(f.byID))
	for _, r := range f.byID {
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeRoleRepo) Update(ctx context.Context, r *domain.Role) error {
	if _, ok := f.byID[r.ID]; !ok {
		return application.ErrNotFound
	}
	f.byID[r.ID] = r
	return nil
}

func (f *fakeRoleRepo) GrantPermission(ctx context.Context, roleID, permissionID string) error {
	r, ok := f.byID[roleID]
	if !ok {
		return application.ErrNotFound
	}
	r.GrantPermission(permissionID)
	return nil
}

func (f *fakeRoleRepo) RevokePermission(ctx context.Context, roleID, permissionID string) error {
	r, ok := f.byID[roleID]
	if !ok {
		return application.ErrNotFound
	}
	r.RevokePermission(permissionID)
	return nil
}

func (f *fakeRoleRepo) Delete(ctx context.Context, id string) error {
	if _, ok := f.byID[id]; !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}

var _ application.RoleRepository = (*fakeRoleRepo)(nil)

type fakeGroupRepo struct {
	byID map[string]*domain.Group
}

func newFakeGroupRepo() *fakeGroupRepo {
	return &fakeGroupRepo{byID: map[string]*domain.Group{}}
}

func (f *fakeGroupRepo) Create(ctx context.Context, g *domain.Group) error {
	f.byID[g.ID] = g
	return nil
}

func (f *fakeGroupRepo) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	g, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return g, nil
}

func (f *fakeGroupRepo) List(ctx context.Context) ([]*domain.Group, error) {
	out := make([]*domain.Group, 0, len(f.byID))
	for _, g := range f.byID {
		out = append(out, g)
	}
	return out, nil
}

func (f *fakeGroupRepo) ListByMemberUserID(ctx context.Context, userID string) ([]*domain.Group, error) {
	out := []*domain.Group{}
	for _, g := range f.byID {
		if g.HasMember(userID) {
			out = append(out, g)
		}
	}
	return out, nil
}

func (f *fakeGroupRepo) Update(ctx context.Context, g *domain.Group) error {
	if _, ok := f.byID[g.ID]; !ok {
		return application.ErrNotFound
	}
	f.byID[g.ID] = g
	return nil
}

func (f *fakeGroupRepo) AddMember(ctx context.Context, groupID, userID string) error {
	g, ok := f.byID[groupID]
	if !ok {
		return application.ErrNotFound
	}
	g.AddMember(userID)
	return nil
}

func (f *fakeGroupRepo) RemoveMember(ctx context.Context, groupID, userID string) error {
	g, ok := f.byID[groupID]
	if !ok {
		return application.ErrNotFound
	}
	g.RemoveMember(userID)
	return nil
}

func (f *fakeGroupRepo) AssignRole(ctx context.Context, groupID, roleID string) error {
	g, ok := f.byID[groupID]
	if !ok {
		return application.ErrNotFound
	}
	g.AssignRole(roleID)
	return nil
}

func (f *fakeGroupRepo) Delete(ctx context.Context, id string) error {
	if _, ok := f.byID[id]; !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}

var _ application.GroupRepository = (*fakeGroupRepo)(nil)

type fakePermissionRepo struct {
	byID map[string]*domain.Permission
}

func newFakePermissionRepo() *fakePermissionRepo {
	return &fakePermissionRepo{byID: map[string]*domain.Permission{}}
}

func (f *fakePermissionRepo) Create(ctx context.Context, p *domain.Permission) error {
	f.byID[p.ID] = p
	return nil
}

func (f *fakePermissionRepo) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	p, ok := f.byID[id]
	if !ok {
		return nil, application.ErrNotFound
	}
	return p, nil
}

func (f *fakePermissionRepo) ListByServiceID(ctx context.Context, serviceID string) ([]*domain.Permission, error) {
	out := []*domain.Permission{}
	for _, p := range f.byID {
		if p.ServiceID == serviceID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakePermissionRepo) GetByServiceIDAndKey(ctx context.Context, serviceID, key string) (*domain.Permission, error) {
	for _, p := range f.byID {
		if p.ServiceID == serviceID && p.Key == key {
			return p, nil
		}
	}
	return nil, application.ErrNotFound
}

func (f *fakePermissionRepo) ListByIDs(ctx context.Context, ids []string) ([]*domain.Permission, error) {
	out := []*domain.Permission{}
	for _, id := range ids {
		if p, ok := f.byID[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakePermissionRepo) Delete(ctx context.Context, id string) error {
	if _, ok := f.byID[id]; !ok {
		return application.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}

var _ application.PermissionRepository = (*fakePermissionRepo)(nil)
