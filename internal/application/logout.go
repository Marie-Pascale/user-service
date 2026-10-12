package application

import "context"

type LogoutUseCase struct {
	Sessions SessionRepository
}

func NewLogoutUseCase(sessions SessionRepository) *LogoutUseCase {
	return &LogoutUseCase{Sessions: sessions}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	return uc.Sessions.Delete(ctx, sessionID)
}
