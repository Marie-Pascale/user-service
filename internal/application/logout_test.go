package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLogoutUseCase_Execute(t *testing.T) {
	sessions := newFakeSessionRepo()
	users := newFakeUserRepo()
	u := mustNewTestUser(t, users, "Alice", "alice@example.com")
	createSession := NewCreateSessionUseCase(sessions, users, time.Hour)
	logout := NewLogoutUseCase(sessions)

	session, _, err := createSession.Execute(context.Background(), u.ID, "203.0.113.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	if err := logout.Execute(context.Background(), session.ID); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	_, err = sessions.GetByTokenHash(context.Background(), session.TokenHash)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error after logout = %v, want %v", err, ErrNotFound)
	}
}
