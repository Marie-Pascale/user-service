//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"user-service/internal/application"
	"user-service/internal/domain"
)

func TestPostgresSessionRepository_CreateAndGetByTokenHash(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	sessionRepo := NewPostgresSessionRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "Session Test", "session-repo-test@example.com")
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = userRepo.Delete(ctx, u.ID) })

	session, plainToken, err := domain.NewSession(domain.NewID(), u.ID, "203.0.113.1", "Mozilla/5.0", time.Hour)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := sessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = sessionRepo.Delete(ctx, session.ID) })

	got, err := sessionRepo.GetByTokenHash(ctx, domain.HashToken(plainToken))
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if got.IPAddress != "203.0.113.1" {
		t.Errorf("IPAddress = %q, want %q", got.IPAddress, "203.0.113.1")
	}
	if got.UserAgent != "Mozilla/5.0" {
		t.Errorf("UserAgent = %q, want %q", got.UserAgent, "Mozilla/5.0")
	}
}

func TestPostgresSessionRepository_GetByTokenHash_NotFound(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	sessionRepo := NewPostgresSessionRepository(pool)

	_, err = sessionRepo.GetByTokenHash(context.Background(), "nonexistent-hash")
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestPostgresSessionRepository_Touch(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	sessionRepo := NewPostgresSessionRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "Touch Test", "session-touch-test@example.com")
	_ = userRepo.Create(ctx, u)
	t.Cleanup(func() { _ = userRepo.Delete(ctx, u.ID) })

	session, _, _ := domain.NewSession(domain.NewID(), u.ID, "203.0.113.1", "Mozilla/5.0", time.Hour)
	_ = sessionRepo.Create(ctx, session)
	t.Cleanup(func() { _ = sessionRepo.Delete(ctx, session.ID) })

	if err := sessionRepo.Touch(ctx, session.ID); err != nil {
		t.Fatalf("Touch: %v", err)
	}

	got, err := sessionRepo.GetByTokenHash(ctx, session.TokenHash)
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if !got.LastUsedAt.After(session.CreatedAt) {
		t.Error("LastUsedAt should have advanced in the database after Touch")
	}
}

func TestPostgresSessionRepository_DeleteAllForUser(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	sessionRepo := NewPostgresSessionRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "DeleteAll Test", "session-deleteall-test@example.com")
	_ = userRepo.Create(ctx, u)
	t.Cleanup(func() { _ = userRepo.Delete(ctx, u.ID) })

	s1, _, _ := domain.NewSession(domain.NewID(), u.ID, "203.0.113.1", "Mozilla/5.0", time.Hour)
	s2, _, _ := domain.NewSession(domain.NewID(), u.ID, "203.0.113.2", "Mozilla/5.0", time.Hour)
	_ = sessionRepo.Create(ctx, s1)
	_ = sessionRepo.Create(ctx, s2)

	if err := sessionRepo.DeleteAllForUser(ctx, u.ID); err != nil {
		t.Fatalf("DeleteAllForUser: %v", err)
	}

	if _, err := sessionRepo.GetByTokenHash(ctx, s1.TokenHash); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("s1 should have been deleted, error = %v", err)
	}
	if _, err := sessionRepo.GetByTokenHash(ctx, s2.TokenHash); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("s2 should have been deleted, error = %v", err)
	}
}
