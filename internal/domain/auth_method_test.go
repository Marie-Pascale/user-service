package domain

import "testing"

// fakeHasher est un PasswordHasher minimal pour les tests du domaine.
// On ne veut PAS dependre de bcrypt ici : le domaine doit rester
// testable sans aucune dependance externe.
type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakeHasher) Verify(password, hash string) bool {
	return "hashed:"+password == hash
}

func TestNewLocalAuthMethod(t *testing.T) {
	hasher := fakeHasher{}

	t.Run("cree une methode locale valide", func(t *testing.T) {
		am, err := NewLocalAuthMethod("1", "user-1", "s3cret", hasher)
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if am.Provider != ProviderLocal {
			t.Errorf("Provider = %q, attendu %q", am.Provider, ProviderLocal)
		}
		if am.PasswordHash != "hashed:s3cret" {
			t.Errorf("PasswordHash = %q, attendu %q", am.PasswordHash, "hashed:s3cret")
		}
	})

	t.Run("rejette un userID vide", func(t *testing.T) {
		_, err := NewLocalAuthMethod("1", "", "s3cret", hasher)
		if err != ErrEmptyUserID {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyUserID)
		}
	})

	t.Run("rejette un mot de passe vide", func(t *testing.T) {
		_, err := NewLocalAuthMethod("1", "user-1", "", hasher)
		if err != ErrEmptyPassword {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyPassword)
		}
	})
}

func TestNewOAuthMethod(t *testing.T) {
	t.Run("cree une methode Google valide", func(t *testing.T) {
		am, err := NewOAuthMethod("1", "user-1", ProviderGoogle, "google-sub-123")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if am.ProviderUserID != "google-sub-123" {
			t.Errorf("ProviderUserID = %q, attendu %q", am.ProviderUserID, "google-sub-123")
		}
	})

	t.Run("rejette un provider invalide", func(t *testing.T) {
		_, err := NewOAuthMethod("1", "user-1", ProviderLocal, "abc")
		if err != ErrInvalidProvider {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidProvider)
		}
	})

	t.Run("rejette un providerUserID vide", func(t *testing.T) {
		_, err := NewOAuthMethod("1", "user-1", ProviderGoogle, "")
		if err != ErrEmptyProviderUserID {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyProviderUserID)
		}
	})
}

func TestAuthMethod_Authenticate(t *testing.T) {
	hasher := fakeHasher{}

	t.Run("mot de passe correct pour une methode locale", func(t *testing.T) {
		am, _ := NewLocalAuthMethod("1", "user-1", "s3cret", hasher)
		if !am.Authenticate("s3cret", hasher) {
			t.Error("Authenticate devrait retourner true pour le bon mot de passe")
		}
	})

	t.Run("mot de passe incorrect pour une methode locale", func(t *testing.T) {
		am, _ := NewLocalAuthMethod("1", "user-1", "s3cret", hasher)
		if am.Authenticate("wrong", hasher) {
			t.Error("Authenticate devrait retourner false pour un mauvais mot de passe")
		}
	})

	t.Run("retourne toujours false pour une methode tierce", func(t *testing.T) {
		am, _ := NewOAuthMethod("1", "user-1", ProviderGoogle, "google-sub-123")
		if am.Authenticate("n'importe quoi", hasher) {
			t.Error("Authenticate devrait retourner false pour une methode OAuth")
		}
	})
}
