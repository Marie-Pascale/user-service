package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

type RegisterUserUseCase struct {
	CreateUser         *CreateUserUseCase
	RegisterAuthMethod *RegisterAuthMethodUseCase
	Users              UserRepository
}

func NewRegisterUserUseCase(createUser *CreateUserUseCase, registerAuthMethod *RegisterAuthMethodUseCase, users UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{CreateUser: createUser, RegisterAuthMethod: registerAuthMethod, Users: users}
}

func (uc *RegisterUserUseCase) ExecuteLocal(ctx context.Context, name, email, password string) (*domain.User, error) {
	u, err := uc.CreateUser.Execute(ctx, name, email)
	if err != nil {
		return nil, err
	}

	if _, err := uc.RegisterAuthMethod.ExecuteLocal(ctx, u.ID, password); err != nil {
		_ = uc.Users.Delete(ctx, u.ID)
		return nil, err
	}

	return u, nil
}

func (uc *RegisterUserUseCase) ExecuteOAuth(ctx context.Context, name, email string, provider domain.Provider, providerUserID string) (*domain.User, error) {
	u, err := uc.CreateUser.Execute(ctx, name, email)
	if err != nil {
		return nil, err
	}

	if _, err := uc.RegisterAuthMethod.ExecuteOAuth(ctx, u.ID, provider, providerUserID); err != nil {
		_ = uc.Users.Delete(ctx, u.ID)
		return nil, err
	}

	return u, nil
}
