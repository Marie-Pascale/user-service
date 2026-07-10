package application

import (
	"context"
	"errors"
	"testing"

	"gestionnaire-usagers/internal/domain"
)

func TestCreateUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("cree un usager valide", func(t *testing.T) {
		uc := NewCreateUserUseCase(newFakeUserRepo())

		u, err := uc.Execute(ctx, "Alice", "alice@example.com")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if u.Email != "alice@example.com" {
			t.Errorf("Email = %q, attendu %q", u.Email, "alice@example.com")
		}

		// Persiste ?
		got, err := uc.Users.GetByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("l'usager devrait etre persiste: %v", err)
		}
		if got.ID != u.ID {
			t.Errorf("ID persiste = %q, attendu %q", got.ID, u.ID)
		}
	})

	t.Run("rejette un courriel deja utilise", func(t *testing.T) {
		repo := newFakeUserRepo()
		uc := NewCreateUserUseCase(repo)

		if _, err := uc.Execute(ctx, "Alice", "alice@example.com"); err != nil {
			t.Fatalf("premiere creation: %v", err)
		}

		_, err := uc.Execute(ctx, "Alice Bis", "alice@example.com")
		if !errors.Is(err, ErrEmailAlreadyUsed) {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmailAlreadyUsed)
		}
	})

	t.Run("propage une erreur de validation du domaine", func(t *testing.T) {
		uc := NewCreateUserUseCase(newFakeUserRepo())

		_, err := uc.Execute(ctx, "", "alice@example.com")
		if !errors.Is(err, domain.ErrEmptyName) {
			t.Errorf("erreur = %v, attendu %v", err, domain.ErrEmptyName)
		}
	})
}
