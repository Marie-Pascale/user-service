package application

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
)

func TestAuthenticateUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T) (*AuthenticateUserUseCase, *domain.User) {
		users := newFakeUserRepo()
		authMethods := newFakeAuthMethodRepo()
		hasher := fakeHasher{}

		u, _ := domain.NewUser(domain.NewID(), "Alice", "alice@example.com")
		_ = users.Create(ctx, u)

		am, err := domain.NewLocalAuthMethod(domain.NewID(), u.ID, "s3cret", hasher)
		if err != nil {
			t.Fatalf("construction de la methode locale: %v", err)
		}
		_ = authMethods.Create(ctx, am)

		return NewAuthenticateUserUseCase(users, authMethods, hasher), u
	}

	t.Run("authentifie avec le bon mot de passe", func(t *testing.T) {
		uc, u := setup(t)

		got, err := uc.Execute(ctx, "alice@example.com", "s3cret")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if got.ID != u.ID {
			t.Errorf("ID = %q, attendu %q", got.ID, u.ID)
		}
	})

	t.Run("rejette un mauvais mot de passe", func(t *testing.T) {
		uc, _ := setup(t)

		_, err := uc.Execute(ctx, "alice@example.com", "mauvais-mdp")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("rejette un courriel inexistant avec la meme erreur (anti-enumeration)", func(t *testing.T) {
		uc, _ := setup(t)

		_, err := uc.Execute(ctx, "inconnu@example.com", "n'importe-quoi")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("rejette un usager suspendu", func(t *testing.T) {
		uc, u := setup(t)
		u.Suspend()

		_, err := uc.Execute(ctx, "alice@example.com", "s3cret")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidCredentials)
		}
	})
}
