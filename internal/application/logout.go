package application

import "context"

// LogoutUseCase revoque une session -- une deconnexion immediate et
// definitive, contrairement a un JWT qui resterait valide jusqu'a son
// expiration naturelle. C'est precisement l'avantage du choix "session
// opaque stockee en base" fait pour ce projet.
type LogoutUseCase struct {
	Sessions SessionRepository
}

func NewLogoutUseCase(sessions SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{Sessions: sessions}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	return uc.Sessions.Delete(ctx, sessionID)
}
