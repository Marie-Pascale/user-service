package application

import (
	"context"
	"testing"
	"time"
)

func TestCreateSessionUseCase_Execute(t *testing.T) {
	sessions := newFakeSessionRepo()
	uc := NewCreateSessionUseCase(sessions, time.Hour)

	session, plainToken, err := uc.Execute(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if plainToken == "" {
		t.Error("le jeton en clair ne devrait pas etre vide")
	}
	if session.UserID != "user-1" {
		t.Errorf("UserID = %q, attendu %q", session.UserID, "user-1")
	}

	// La session doit avoir ete persistee.
	got, err := sessions.GetByTokenHash(context.Background(), session.TokenHash)
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if got.ID != session.ID {
		t.Errorf("session persistee = %q, attendu %q", got.ID, session.ID)
	}
}
