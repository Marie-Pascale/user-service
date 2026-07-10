package application

import (
	"context"
	"errors"

	"gestionnaire-usagers/internal/domain"
)

var ErrEmailAlreadyUsed = errors.New("email already used by another account")

// CreateUserUseCase orchestre la creation d'un nouvel usager : construction
// et validation via le domaine, verification de l'unicite du courriel,
// puis persistance.
//
// Note : ce type ne connait que l'interface UserRepository, jamais
// PostgresUserRepository -- voir la discussion sur l'inversion de
// dependance plus tot dans le projet.
type CreateUserUseCase struct {
	Users UserRepository
}

func NewCreateUserUseCase(users UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{Users: users}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, name, email string) (*domain.User, error) {
	// La construction du domaine valide deja le format (nom non vide,
	// courriel bien forme) -- inutile de dupliquer cette logique ici.
	u, err := domain.NewUser(domain.NewID(), name, email)
	if err != nil {
		return nil, err
	}

	// L'unicite du courriel, elle, ne peut PAS etre validee par le domaine
	// seul : ca depend de ce qui existe deja en base, une information que
	// seul le repository possede. C'est une regle d'orchestration, donc
	// elle vit ici, dans le cas d'usage -- pas dans domain.NewUser.
	_, err = uc.Users.GetByEmail(ctx, u.Email)
	if err == nil {
		return nil, ErrEmailAlreadyUsed
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err // une vraie erreur technique, pas juste "non trouve"
	}

	if err := uc.Users.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
