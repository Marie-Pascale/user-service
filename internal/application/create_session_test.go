package application

import (
	"context"
	"testing"
	"time"
)

func TestCreateSessionUseCase_Execute(t *testing.T) {
	sessions := newFakeSessionRepo()
	users := newFakeUserRepo()
	u := mustNewTestUser(t, users, "Alice", "alice@example.com")

	uc := NewCreateSessionUseCase(sessions, users, time.Hour)

	session, plainToken, err := uc.Execute(context.Background(), u.ID, "203.0.113.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if plainToken == "" {
		t.Error("le jeton en clair ne devrait pas etre vide")
	}
	if session.UserID != u.ID {
		t.Errorf("UserID = %q, attendu %q", session.UserID, u.ID)
	}

	// La session doit avoir ete persistee.
	got, err := sessions.GetByTokenHash(context.Background(), session.TokenHash)
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("session persistee = %q, attendu %q", got.ID, session.ID)
	}

	// RecordLogin doit avoir ete appele -- voir domain.User.RecordLogin.
	updated, err := users.GetByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if updated.LastLoginAt == nil {
		t.Error("LastLoginAt devrait etre defini apres CreateSessionUseCase.Execute")
	}
}

func TestCreateSessionUseCase_Execute_IgnoresRecordLoginFailure(t *testing.T) {
	// Un usager inexistant ne devrait PAS faire echouer la creation de
	// session -- RecordLogin est best-effort (voir le commentaire dans
	// create_session.go).
	sessions := newFakeSessionRepo()
	users := newFakeUserRepo()
	uc := NewCreateSessionUseCase(sessions, users, time.Hour)

	_, _, err := uc.Execute(context.Background(), "usager-inexistant", "203.0.113.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("erreur inattendue: %v (RecordLogin devrait etre best-effort)", err)
	}
}
