package application

import (
	"context"

	"user-service/internal/domain"
)

type RegisterAuthMethodUseCase struct {
	Users       UserRepository
	AuthMethods AuthMethodRepository
	Hasher      domain.PasswordHasher
}

func NewRegisterAuthMethodUseCase(users UserRepository, authMethods AuthMethodRepository, hasher domain.PasswordHasher) *RegisterAuthMethodUseCase {
	return &RegisterAuthMethodUseCase{Users: users, AuthMethods: authMethods, Hasher: hasher}
}

func (uc *RegisterAuthMethodUseCase) ExecuteLocal(ctx context.Context, userID, password string) (*domain.AuthMethod, error) {
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

func (uc *RegisterAuthMethodUseCase) ExecuteOAuth(ctx context.Context, userID string, provider domain.Provider, providerUserID string) (*domain.AuthMethod, error) {
	if _, err := uc.Users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

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
