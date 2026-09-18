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
	Users    UserRepository
	TTL      time.Duration
}

func NewCreateSessionUseCase(sessions SessionRepository, users UserRepository, ttl time.Duration) *CreateSessionUseCase {
	return &CreateSessionUseCase{Sessions: sessions, Users: users, TTL: ttl}
}

// Execute cree une session pour l'usager donne. Retourne le jeton en
// clair -- a transmettre au client (ex: dans le corps de la reponse HTTP
// de login) -- qui devra le presenter pour chaque requete authentifiee
// suivante.
//
// ipAddress et userAgent viennent de la requete HTTP de connexion (voir
// AuthHandler.Login) -- captures UNE SEULE FOIS ici, a la creation, pour
// servir d'ancre a Session.IsSuspicious par la suite.
//
// Enregistre aussi last_login_at ICI plutot que dans AuthenticateUserUseCase
// -- voir domain.User.RecordLogin pour le raisonnement complet (couvrir
// uniformement tous les futurs fournisseurs d'authentification, pas
// seulement le mot de passe local).
func (uc *CreateSessionUseCase) Execute(ctx context.Context, userID, ipAddress, userAgent string) (session *domain.Session, plainToken string, err error) {
	session, plainToken, err = domain.NewSession(domain.NewID(), userID, ipAddress, userAgent, uc.TTL)
	if err != nil {
		return nil, "", err
	}

	if err := uc.Sessions.Create(ctx, session); err != nil {
		return nil, "", err
	}

	// Best-effort : un echec de RecordLogin ne doit pas empecher la
	// connexion elle-meme -- la session est deja creee et valide a ce
	// stade, last_login_at n'est qu'une metadonnee informative, pas une
	// condition de securite. Meme logique que UserAuthMiddleware avec
	// Sessions.Touch.
	_ = uc.Users.RecordLogin(ctx, userID)

	return session, plainToken, nil
}
