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
