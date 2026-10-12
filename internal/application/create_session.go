package application

import (
	"context"
	"time"

	"user-service/internal/domain"
)

type CreateSessionUseCase struct {
	Sessions SessionRepository
	Users    UserRepository
	TTL      time.Duration
}

func NewCreateSessionUseCase(sessions SessionRepository, users UserRepository, ttl time.Duration) *CreateSessionUseCase {
	return &CreateSessionUseCase{Sessions: sessions, Users: users, TTL: ttl}
}

func (uc *CreateSessionUseCase) Execute(ctx context.Context, userID, ipAddress, userAgent string) (session *domain.Session, plainToken string, err error) {
	session, plainToken, err = domain.NewSession(domain.NewID(), userID, ipAddress, userAgent, uc.TTL)
	if err != nil {
		return nil, "", err
	}

	if err := uc.Sessions.Create(ctx, session); err != nil {
		return nil, "", err
	}

	_ = uc.Users.RecordLogin(ctx, userID)

	return session, plainToken, nil
}
