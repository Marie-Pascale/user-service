package application

import (
	"context"
	"errors"

	"user-service/internal/domain"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthenticateUserUseCase struct {
	Users       UserRepository
	AuthMethods AuthMethodRepository
	Hasher      domain.PasswordHasher
}

func NewAuthenticateUserUseCase(users UserRepository, authMethods AuthMethodRepository, hasher domain.PasswordHasher) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{Users: users, AuthMethods: authMethods, Hasher: hasher}
}

func (uc *AuthenticateUserUseCase) Execute(ctx context.Context, email, password string) (*domain.User, error) {
	u, err := uc.Users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.IsActive() {
		return nil, ErrInvalidCredentials
	}

	methods, err := uc.AuthMethods.ListByUserID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	for _, am := range methods {
		if am.Provider != domain.ProviderLocal {
			continue
		}
		if am.Authenticate(password, uc.Hasher) {
			_ = uc.AuthMethods.MarkUsed(ctx, am.ID)
			return u, nil
		}
	}

	return nil, ErrInvalidCredentials
}
