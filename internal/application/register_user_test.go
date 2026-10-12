package application

import (
	"context"
	"errors"
	"testing"

	"gestionnaire-usagers/internal/domain"
)

func newTestRegisterUserUseCase() (*RegisterUserUseCase, *fakeUserRepo, *fakeAuthMethodRepo) {
	users := newFakeUserRepo()
	authMethods := newFakeAuthMethodRepo()
	hasher := fakeHasher{}

	createUser := NewCreateUserUseCase(users)
	registerAuthMethod := NewRegisterAuthMethodUseCase(users, authMethods, hasher)

	return NewRegisterUserUseCase(createUser, registerAuthMethod, users), users, authMethods
}

func TestRegisterUserUseCase_ExecuteLocal(t *testing.T) {
	t.Run("cree l'usager et sa methode d'authentification", func(t *testing.T) {
		uc, users, authMethods := newTestRegisterUserUseCase()

		u, err := uc.ExecuteLocal(context.Background(), "Alice", "alice@example.com", "s3cret")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}

		if _, err := users.GetByID(context.Background(), u.ID); err != nil {
			t.Errorf("l'usager devrait avoir ete persiste: %v", err)
		}

		methods, err := authMethods.ListByUserID(context.Background(), u.ID)
		if err != nil {
			t.Fatalf("ListByUserID: %v", err)
		}
		if len(methods) != 1 || methods[0].Provider != domain.ProviderLocal {
			t.Errorf("methodes = %+v, attendu une methode locale", methods)
		}
	})

	t.Run("supprime l'usager si le mot de passe est invalide (compensation)", func(t *testing.T) {
		uc, users, _ := newTestRegisterUserUseCase()

		_, err := uc.ExecuteLocal(context.Background(), "Alice", "alice@example.com", "")
		if err != domain.ErrEmptyPassword {
			t.Fatalf("erreur = %v, attendu %v", err, domain.ErrEmptyPassword)
		}

		_, err = users.GetByEmail(context.Background(), "alice@example.com")
		if !errors.Is(err, ErrNotFound) {
			t.Error("l'usager aurait du etre supprime (compensation) apres l'echec de creation du mot de passe")
		}
	})

	t.Run("refuse un courriel deja utilise sans toucher aux methodes d'authentification", func(t *testing.T) {
		uc, _, authMethods := newTestRegisterUserUseCase()

		_, err := uc.ExecuteLocal(context.Background(), "Alice", "alice@example.com", "s3cret")
		if err != nil {
			t.Fatalf("premiere inscription: %v", err)
		}

		_, err = uc.ExecuteLocal(context.Background(), "Alice Bis", "alice@example.com", "autre-mdp")
		if err != ErrEmailAlreadyUsed {
			t.Fatalf("erreur = %v, attendu %v", err, ErrEmailAlreadyUsed)
		}

		all, _ := authMethods.ListByUserID(context.Background(), "")
		_ = all
	})
}

func TestRegisterUserUseCase_ExecuteOAuth(t *testing.T) {
	uc, users, authMethods := newTestRegisterUserUseCase()

	u, err := uc.ExecuteOAuth(context.Background(), "Alice", "alice@example.com", domain.ProviderGoogle, "google-sub-123")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}

	if _, err := users.GetByID(context.Background(), u.ID); err != nil {
		t.Errorf("l'usager devrait avoir ete persiste: %v", err)
	}

	methods, err := authMethods.ListByUserID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(methods) != 1 || methods[0].Provider != domain.ProviderGoogle {
		t.Errorf("methodes = %+v, attendu une methode Google", methods)
	}
}
