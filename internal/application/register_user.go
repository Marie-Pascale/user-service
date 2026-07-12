package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// RegisterUserUseCase orchestre la creation d'un usager ET de sa toute
// premiere methode d'authentification, en une seule operation cote client
// -- plutot que deux appels HTTP separes (POST /users, puis
// POST /users/{id}/password).
//
// Compose CreateUserUseCase et RegisterAuthMethodUseCase plutot que de
// dupliquer leur logique de validation -- les deux restent utilisables
// independamment (ex: RegisterAuthMethodUseCase sert aussi a AJOUTER une
// deuxieme methode d'authentification a un compte existant).
//
// LIMITE CONNUE : ces deux etapes ne sont PAS dans une transaction de base
// de donnees unique -- les repositories ne partagent pas de connexion
// entre eux actuellement. Si la creation de la methode d'authentification
// echoue apres que l'usager ait deja ete cree, on tente une suppression de
// compensation plutot que de laisser un usager fantome sans aucun moyen de
// se connecter. Ce n'est qu'un "best effort" (la suppression elle-meme
// pourrait echouer) -- une vraie solution robuste utiliserait une
// transaction partagee (Unit of Work), a introduire si ce genre
// d'incoherence devient un probleme reel en pratique.
type RegisterUserUseCase struct {
	CreateUser         *CreateUserUseCase
	RegisterAuthMethod *RegisterAuthMethodUseCase
	Users              UserRepository
}

func NewRegisterUserUseCase(createUser *CreateUserUseCase, registerAuthMethod *RegisterAuthMethodUseCase, users UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{CreateUser: createUser, RegisterAuthMethod: registerAuthMethod, Users: users}
}

// ExecuteLocal cree un usager avec un mot de passe local.
func (uc *RegisterUserUseCase) ExecuteLocal(ctx context.Context, name, email, password string) (*domain.User, error) {
	u, err := uc.CreateUser.Execute(ctx, name, email)
	if err != nil {
		return nil, err
	}

	if _, err := uc.RegisterAuthMethod.ExecuteLocal(ctx, u.ID, password); err != nil {
		_ = uc.Users.Delete(ctx, u.ID) // compensation -- voir la doc du type
		return nil, err
	}

	return u, nil
}

// ExecuteOAuth cree un usager authentifie via un fournisseur tiers
// (Google, Facebook...).
//
// IMPORTANT : cette methode ne verifie RIEN par elle-meme -- elle fait
// confiance a providerUserID tel quel. Elle suppose qu'un flot
// d'autorisation OAuth reel (echange de code, verification de jeton
// signe aupres du fournisseur) a deja eu lieu EN AMONT, et que
// providerUserID en est le resultat verifie. Ce flot n'existe pas encore
// dans ce projet -- c'est pourquoi cette methode n'est PAS branchee sur
// un handler HTTP public pour l'instant (voir UserHandler.Create). Elle
// existe pour que la structure soit prete le jour ou ce flot sera
// implemente, sans avoir a retoucher le domaine ou l'orchestration.
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
