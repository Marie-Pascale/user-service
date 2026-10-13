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
			t.Fatalf("create local auth method: %v", err)
		}
		_ = authMethods.Create(ctx, am)

		return NewAuthenticateUserUseCase(users, authMethods, hasher), u
	}

	t.Run("authenticates with the correct password", func(t *testing.T) {
		uc, u := setup(t)

		got, err := uc.Execute(ctx, "alice@example.com", "s3cret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != u.ID {
			t.Errorf("ID = %q, want %q", got.ID, u.ID)
		}
	})

	t.Run("rejects an incorrect password", func(t *testing.T) {
		uc, _ := setup(t)

		_, err := uc.Execute(ctx, "alice@example.com", "incorrect-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("error = %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("rejects a nonexistent email address with the same error (anti-enumeration)", func(t *testing.T) {
		uc, _ := setup(t)

		_, err := uc.Execute(ctx, "unknown@example.com", "n'importe-quoi")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("error = %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("rejects a suspended user", func(t *testing.T) {
		uc, u := setup(t)
		u.Suspend()

		_, err := uc.Execute(ctx, "alice@example.com", "s3cret")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("error = %v, want %v", err, ErrInvalidCredentials)
		}
	})
}
