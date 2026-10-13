package application

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
)

func TestCreateUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("creates a valid user", func(t *testing.T) {
		uc := NewCreateUserUseCase(newFakeUserRepo())

		u, err := uc.Execute(ctx, "Alice", "alice@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Email != "alice@example.com" {
			t.Errorf("Email = %q, want %q", u.Email, "alice@example.com")
		}

		got, err := uc.Users.GetByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("the user should have been persisted: %v", err)
		}
		if got.ID != u.ID {
			t.Errorf("persisted ID = %q, want %q", got.ID, u.ID)
		}
	})

	t.Run("rejects an email address that is already in use", func(t *testing.T) {
		repo := newFakeUserRepo()
		uc := NewCreateUserUseCase(repo)

		if _, err := uc.Execute(ctx, "Alice", "alice@example.com"); err != nil {
			t.Fatalf("first creation: %v", err)
		}

		_, err := uc.Execute(ctx, "Alice Bis", "alice@example.com")
		if !errors.Is(err, ErrEmailAlreadyUsed) {
			t.Errorf("error = %v, want %v", err, ErrEmailAlreadyUsed)
		}
	})

	t.Run("propagates a domain validation error", func(t *testing.T) {
		uc := NewCreateUserUseCase(newFakeUserRepo())

		_, err := uc.Execute(ctx, "", "alice@example.com")
		if !errors.Is(err, domain.ErrEmptyName) {
			t.Errorf("error = %v, want %v", err, domain.ErrEmptyName)
		}
	})
}
