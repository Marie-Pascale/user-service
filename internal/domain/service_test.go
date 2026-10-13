package domain

import (
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	t.Run("creates a valid service with an API key", func(t *testing.T) {
		svc, plainKey, err := NewService("1", "orders-service", "gere les commandes")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if svc.Name != "orders-service" {
			t.Errorf("Name = %q, want %q", svc.Name, "orders-service")
		}
		if plainKey == "" {
			t.Error("the plaintext API key should not be empty")
		}
		if svc.APIKeyHash == "" {
			t.Error("APIKeyHash should not be empty")
		}
		if svc.APIKeyHash == plainKey {
			t.Error("APIKeyHash should never equal the plaintext key")
		}
		if svc.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
		if svc.CreatedAt != svc.UpdatedAt {
			t.Error("CreatedAt and UpdatedAt should be identical at creation")
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		_, _, err := NewService("1", "", "description")
		if err != ErrEmptyServiceName {
			t.Errorf("error = %v, want %v", err, ErrEmptyServiceName)
		}
	})

	t.Run("generates a different key on each call", func(t *testing.T) {
		_, key1, _ := NewService("1", "svc-a", "")
		_, key2, _ := NewService("2", "svc-b", "")
		if key1 == key2 {
			t.Error("two services should never receive the same API key")
		}
	})
}

func TestService_VerifyAPIKey(t *testing.T) {
	svc, plainKey, _ := NewService("1", "orders-service", "")

	t.Run("accepts the correct key", func(t *testing.T) {
		if !svc.VerifyAPIKey(plainKey) {
			t.Error("VerifyAPIKey should return true for the correct key")
		}
	})

	t.Run("rejects an incorrect key", func(t *testing.T) {
		if svc.VerifyAPIKey("incorrect-key") {
			t.Error("VerifyAPIKey should return false for an incorrect key")
		}
	})
}

func TestService_RotateAPIKey(t *testing.T) {
	svc, oldKey, _ := NewService("1", "orders-service", "")
	originalUpdatedAt := svc.UpdatedAt

	time.Sleep(time.Millisecond)
	newKey, err := svc.RotateAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newKey == oldKey {
		t.Error("the new key should not be identical to the old key")
	}
	if svc.VerifyAPIKey(oldKey) {
		t.Error("the old key should no longer be valid after rotation")
	}
	if !svc.VerifyAPIKey(newKey) {
		t.Error("the new key should be valid after rotation")
	}
	if !svc.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should advance after RotateAPIKey")
	}
}

func TestService_UpdateDetails(t *testing.T) {
	t.Run("updates name and description without changing the key", func(t *testing.T) {
		svc, plainKey, _ := NewService("1", "orders-service", "old description")
		hashBefore := svc.APIKeyHash

		err := svc.UpdateDetails("orders-service-v2", "new description")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if svc.Name != "orders-service-v2" {
			t.Errorf("Name = %q, want %q", svc.Name, "orders-service-v2")
		}
		if svc.APIKeyHash != hashBefore {
			t.Error("APIKeyHash should not change after UpdateDetails")
		}
		if !svc.VerifyAPIKey(plainKey) {
			t.Error("the existing key should remain valid after UpdateDetails")
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		svc, _, _ := NewService("1", "orders-service", "")

		err := svc.UpdateDetails("", "description")
		if err != ErrEmptyServiceName {
			t.Errorf("error = %v, want %v", err, ErrEmptyServiceName)
		}
	})
}
