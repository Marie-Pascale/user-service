package httpapi

import (
	"context"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// fakeUserRepo est une implementation en memoire de application.UserRepository,
// utilisee uniquement dans les tests de ce package (suffixe _test.go).
// Duplique volontairement le fake equivalent dans internal/application --
// les tests de packages differents ne peuvent pas partager de code
// non-exporte en Go.
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

var _ application.UserRepository = (*fakeUserRepo)(nil)

// fakeServiceRepo est une implementation en memoire de
// application.ServiceRepository, pour tester le middleware sans base
// de donnees.
type fakeServiceRepo struct {
	byID       map[string]*domain.Service
	byKeyHash  map[string]*domain.Service
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

func (f *fakeServiceRepo) UpdateAPIKeyHash(ctx context.Context, id, newHash string) error {
	s, ok := f.byID[id]
	if !ok {
		return application.ErrNotFound
	}
	delete(f.byKeyHash, s.APIKeyHash)
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

// fakeSessionRepo est une implementation en memoire de
// application.SessionRepository, pour tester UserAuthMiddleware sans
// base de donnees.
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
