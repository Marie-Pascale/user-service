package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		userName string
		email    string
		wantErr  error
	}{
		{
			name:     "usager valide",
			id:       "1",
			userName: "Alice",
			email:    "alice@example.com",
			wantErr:  nil,
		},
		{
			name:     "nom vide",
			id:       "2",
			userName: "",
			email:    "bob@example.com",
			wantErr:  ErrEmptyName,
		},
		{
			name:     "nom avec seulement des espaces",
			id:       "3",
			userName: "   ",
			email:    "bob@example.com",
			wantErr:  ErrEmptyName,
		},
		{
			name:     "courriel sans @",
			id:       "4",
			userName: "Bob",
			email:    "bob-example.com",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "courriel sans domaine",
			id:       "5",
			userName: "Bob",
			email:    "bob@",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "courriel sans point dans le domaine",
			id:       "6",
			userName: "Bob",
			email:    "bob@examplecom",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "courriel normalise en minuscules",
			id:       "7",
			userName: "Carole",
			email:    "CAROLE@EXAMPLE.COM",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.id, tt.userName, tt.email)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("erreur = %v, attendu %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if u.ID != tt.id {
				t.Errorf("ID = %q, attendu %q", u.ID, tt.id)
			}
			if u.Status != StatusActive {
				t.Errorf("Status = %q, attendu %q (un nouvel usager doit etre actif)", u.Status, StatusActive)
			}
			if u.CreatedAt.IsZero() {
				t.Error("CreatedAt ne devrait pas etre a zero")
			}
			if u.CreatedAt != u.UpdatedAt {
				t.Error("CreatedAt et UpdatedAt devraient etre identiques a la creation")
			}
			if u.RoleIDs == nil || u.PermissionIDs == nil {
				t.Error("RoleIDs et PermissionIDs devraient etre initialises (slices vides, pas nil)")
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	if !u.IsActive() {
		t.Error("un nouvel usager devrait etre actif par defaut")
	}

	u.Suspend()

	if u.IsActive() {
		t.Error("un usager suspendu ne devrait plus etre actif")
	}
}

func TestUser_AssignRole(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	u.AssignRole("role-admin")
	if len(u.RoleIDs) != 1 || u.RoleIDs[0] != "role-admin" {
		t.Fatalf("RoleIDs = %v, attendu [\"role-admin\"]", u.RoleIDs)
	}

	// Assigner le meme role deux fois ne doit pas creer de doublon.
	u.AssignRole("role-admin")
	if len(u.RoleIDs) != 1 {
		t.Errorf("RoleIDs = %v, un role duplique ne devrait pas etre ajoute", u.RoleIDs)
	}
}

func TestUser_AssignPermission(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	u.AssignPermission("perm-orders-read")
	if len(u.PermissionIDs) != 1 || u.PermissionIDs[0] != "perm-orders-read" {
		t.Fatalf("PermissionIDs = %v, attendu [\"perm-orders-read\"]", u.PermissionIDs)
	}

	u.AssignPermission("perm-orders-read")
	if len(u.PermissionIDs) != 1 {
		t.Errorf("PermissionIDs = %v, une permission dupliquee ne devrait pas etre ajoutee", u.PermissionIDs)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	t.Run("met a jour nom et courriel valides", func(t *testing.T) {
		u, _ := NewUser("1", "Alice", "alice@example.com")
		u.AssignRole("role-1") // pour verifier que ca survit a la mise a jour

		err := u.UpdateProfile("Alice Dupont", "ALICE.DUPONT@EXAMPLE.COM")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if u.Name != "Alice Dupont" {
			t.Errorf("Name = %q, attendu %q", u.Name, "Alice Dupont")
		}
		if u.Email != "alice.dupont@example.com" {
			t.Errorf("Email = %q, attendu la version en minuscules", u.Email)
		}
		if len(u.RoleIDs) != 1 {
			t.Error("RoleIDs ne devrait pas etre affecte par UpdateProfile")
		}
	})

	t.Run("rejette un nom vide sans modifier l'usager", func(t *testing.T) {
		u, _ := NewUser("1", "Alice", "alice@example.com")

		err := u.UpdateProfile("", "nouveau@example.com")
		if err != ErrEmptyName {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyName)
		}
		if u.Name != "Alice" {
			t.Error("Name ne devrait pas avoir change apres une validation echouee")
		}
	})

	t.Run("rejette un courriel invalide", func(t *testing.T) {
		u, _ := NewUser("1", "Alice", "alice@example.com")

		err := u.UpdateProfile("Alice", "pas-un-courriel")
		if err != ErrInvalidEmail {
			t.Errorf("erreur = %v, attendu %v", err, ErrInvalidEmail)
		}
	})
}

func TestUser_RecordLogin(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	if u.LastLoginAt != nil {
		t.Error("LastLoginAt devrait etre nil avant toute connexion")
	}

	originalUpdatedAt := u.UpdatedAt
	time.Sleep(time.Millisecond)
	u.RecordLogin()

	if u.LastLoginAt == nil {
		t.Fatal("LastLoginAt ne devrait plus etre nil apres RecordLogin")
	}
	if !u.LastLoginAt.After(originalUpdatedAt) {
		t.Error("LastLoginAt devrait etre posterieur a la creation de l'usager")
	}
	if !u.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt devrait aussi avancer apres RecordLogin")
	}
}
