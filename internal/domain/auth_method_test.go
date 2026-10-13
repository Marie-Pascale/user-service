package domain

import "testing"

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakeHasher) Verify(password, hash string) bool {
	return "hashed:"+password == hash
}

func TestNewLocalAuthMethod(t *testing.T) {
	hasher := fakeHasher{}

	t.Run("creates a valid local auth method", func(t *testing.T) {
		am, err := NewLocalAuthMethod("1", "user-1", "s3cret", hasher)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if am.Provider != ProviderLocal {
			t.Errorf("Provider = %q, want %q", am.Provider, ProviderLocal)
		}
		if am.PasswordHash != "hashed:s3cret" {
			t.Errorf("PasswordHash = %q, want %q", am.PasswordHash, "hashed:s3cret")
		}
	})

	t.Run("rejects an empty userID", func(t *testing.T) {
		_, err := NewLocalAuthMethod("1", "", "s3cret", hasher)
		if err != ErrEmptyUserID {
			t.Errorf("error = %v, want %v", err, ErrEmptyUserID)
		}
	})

	t.Run("rejects an empty password", func(t *testing.T) {
		_, err := NewLocalAuthMethod("1", "user-1", "", hasher)
		if err != ErrEmptyPassword {
			t.Errorf("error = %v, want %v", err, ErrEmptyPassword)
		}
	})
}

func TestNewOAuthMethod(t *testing.T) {
	t.Run("creates a valid Google auth method", func(t *testing.T) {
		am, err := NewOAuthMethod("1", "user-1", ProviderGoogle, "google-sub-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if am.ProviderUserID != "google-sub-123" {
			t.Errorf("ProviderUserID = %q, want %q", am.ProviderUserID, "google-sub-123")
		}
	})

	t.Run("rejects an invalid provider", func(t *testing.T) {
		_, err := NewOAuthMethod("1", "user-1", ProviderLocal, "abc")
		if err != ErrInvalidProvider {
			t.Errorf("error = %v, want %v", err, ErrInvalidProvider)
		}
	})

	t.Run("rejects an empty providerUserID", func(t *testing.T) {
		_, err := NewOAuthMethod("1", "user-1", ProviderGoogle, "")
		if err != ErrEmptyProviderUserID {
			t.Errorf("error = %v, want %v", err, ErrEmptyProviderUserID)
		}
	})
}

func TestAuthMethod_Authenticate(t *testing.T) {
	hasher := fakeHasher{}

	t.Run("correct password for a local method", func(t *testing.T) {
		am, _ := NewLocalAuthMethod("1", "user-1", "s3cret", hasher)
		if !am.Authenticate("s3cret", hasher) {
			t.Error("Authenticate should return true for the correct password")
		}
	})

	t.Run("incorrect password for a local method", func(t *testing.T) {
		am, _ := NewLocalAuthMethod("1", "user-1", "s3cret", hasher)
		if am.Authenticate("wrong", hasher) {
			t.Error("Authenticate should return false for an incorrect password")
		}
	})

	t.Run("always returns false for a third-party method", func(t *testing.T) {
		am, _ := NewOAuthMethod("1", "user-1", ProviderGoogle, "google-sub-123")
		if am.Authenticate("n'importe quoi", hasher) {
			t.Error("Authenticate should return false for an OAuth method")
		}
	})
}
