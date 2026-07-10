package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// RegisterAuthMethodUseCase orchestre l'ajout d'une methode d'authentification
// a un usager existant. Un meme usager peut en avoir plusieurs -- voir
// domain.AuthMethod.
type RegisterAuthMethodUseCase struct {
	Users       UserRepository
	AuthMethods AuthMethodRepository
	Hasher      domain.PasswordHasher
}

func NewRegisterAuthMethodUseCase(users UserRepository, authMethods AuthMethodRepository, hasher domain.PasswordHasher) *RegisterAuthMethodUseCase {
	return &RegisterAuthMethodUseCase{Users: users, AuthMethods: authMethods, Hasher: hasher}
}

// ExecuteLocal ajoute une methode d'authentification par mot de passe.
func (uc *RegisterAuthMethodUseCase) ExecuteLocal(ctx context.Context, userID, password string) (*domain.AuthMethod, error) {
	// On verifie que l'usager existe avant d'aller plus loin -- sinon on
	// se retrouverait avec une methode d'authentification pointant vers
	// un usager fantome.
	if _, err := uc.Users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	am, err := domain.NewLocalAuthMethod(domain.NewID(), userID, password, uc.Hasher)
	if err != nil {
		return nil, err
	}

	if err := uc.AuthMethods.Create(ctx, am); err != nil {
		return nil, err
	}
	return am, nil
}

// ExecuteOAuth ajoute une methode d'authentification tierce (Google, Facebook).
func (uc *RegisterAuthMethodUseCase) ExecuteOAuth(ctx context.Context, userID string, provider domain.Provider, providerUserID string) (*domain.AuthMethod, error) {
	if _, err := uc.Users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	// Un meme compte tiers ne devrait pas pouvoir se lier a deux usagers
	// differents -- la contrainte UNIQUE(provider, provider_user_id) le
	// garantit deja en base, mais on verifie ici aussi pour retourner une
	// erreur claire plutot qu'une erreur SQL brute.
	if existing, err := uc.AuthMethods.GetByProviderUserID(ctx, provider, providerUserID); err == nil && existing.UserID != userID {
		return nil, ErrProviderAccountAlreadyLinked
	}

	am, err := domain.NewOAuthMethod(domain.NewID(), userID, provider, providerUserID)
	if err != nil {
		return nil, err
	}

	if err := uc.AuthMethods.Create(ctx, am); err != nil {
		return nil, err
	}
	return am, nil
}
