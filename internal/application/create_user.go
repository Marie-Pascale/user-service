package application

import (
	"context"
	"errors"

	"gestionnaire-usagers/internal/domain"
)

var ErrEmailAlreadyUsed = errors.New("email already used by another account")

type CreateUserUseCase struct {
	Users UserRepository
}

func NewCreateUserUseCase(users UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{Users: users}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, name, email string) (*domain.User, error) {
	u, err := domain.NewUser(domain.NewID(), name, email)
	if err != nil {
		return nil, err
	}

	_, err = uc.Users.GetByEmail(ctx, u.Email)
	if err == nil {
		return nil, ErrEmailAlreadyUsed
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if err := uc.Users.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
