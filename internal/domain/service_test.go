package domain

import "testing"

func TestNewService(t *testing.T) {
	t.Run("cree un service valide avec une cle API", func(t *testing.T) {
		svc, plainKey, err := NewService("1", "orders-service", "gere les commandes")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if svc.Name != "orders-service" {
			t.Errorf("Name = %q, attendu %q", svc.Name, "orders-service")
		}
		if plainKey == "" {
			t.Error("la cle API en clair ne devrait pas etre vide")
		}
		if svc.APIKeyHash == "" {
			t.Error("APIKeyHash ne devrait pas etre vide")
		}
		if svc.APIKeyHash == plainKey {
			t.Error("APIKeyHash ne devrait jamais etre egal a la cle en clair")
		}
	})

	t.Run("rejette un nom vide", func(t *testing.T) {
		_, _, err := NewService("1", "", "description")
		if err != ErrEmptyServiceName {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyServiceName)
		}
	})

	t.Run("genere une cle differente a chaque appel", func(t *testing.T) {
		_, key1, _ := NewService("1", "svc-a", "")
		_, key2, _ := NewService("2", "svc-b", "")
		if key1 == key2 {
			t.Error("deux services ne devraient jamais recevoir la meme cle API")
		}
	})
}

func TestService_VerifyAPIKey(t *testing.T) {
	svc, plainKey, _ := NewService("1", "orders-service", "")

	t.Run("accepte la bonne cle", func(t *testing.T) {
		if !svc.VerifyAPIKey(plainKey) {
			t.Error("VerifyAPIKey devrait retourner true pour la bonne cle")
		}
	})

	t.Run("rejette une mauvaise cle", func(t *testing.T) {
		if svc.VerifyAPIKey("mauvaise-cle") {
			t.Error("VerifyAPIKey devrait retourner false pour une mauvaise cle")
		}
	})
}

func TestService_RotateAPIKey(t *testing.T) {
	svc, oldKey, _ := NewService("1", "orders-service", "")

	newKey, err := svc.RotateAPIKey()
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}

	if newKey == oldKey {
		t.Error("la nouvelle cle ne devrait pas etre identique a l'ancienne")
	}
	if svc.VerifyAPIKey(oldKey) {
		t.Error("l'ancienne cle ne devrait plus etre valide apres rotation")
	}
	if !svc.VerifyAPIKey(newKey) {
		t.Error("la nouvelle cle devrait etre valide apres rotation")
	}
}
