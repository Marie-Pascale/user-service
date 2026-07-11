package application

import (
	"context"
	"time"

	"gestionnaire-usagers/internal/domain"
)

// CreateSessionUseCase cree une session apres une authentification
// reussie. Separe d'AuthenticateUserUseCase (qui verifie SEULEMENT les
// identifiants) parce que ce sont deux responsabilites distinctes : l'une
// verifie "est-ce que ce mot de passe est correct", l'autre "je veux
// maintenant garder cet usager connecte". Un handler HTTP (Login) appelle
// les deux en sequence.
type CreateSessionUseCase struct {
	Sessions SessionRepository
	TTL      time.Duration
}

func NewCreateSessionUseCase(sessions SessionRepository, ttl time.Duration) *CreateSessionUseCase {
	return &CreateSessionUseCase{Sessions: sessions, TTL: ttl}
}

// Execute cree une session pour l'usager donne. Retourne le jeton en
// clair -- a transmettre au client (ex: dans le corps de la reponse HTTP
// de login) -- qui devra le presenter pour chaque requete authentifiee
// suivante.
func (uc *CreateSessionUseCase) Execute(ctx context.Context, userID string) (session *domain.Session, plainToken string, err error) {
	session, plainToken, err = domain.NewSession(domain.NewID(), userID, uc.TTL)
	if err != nil {
		return nil, "", err
	}

	if err := uc.Sessions.Create(ctx, session); err != nil {
		return nil, "", err
	}

	return session, plainToken, nil
}
