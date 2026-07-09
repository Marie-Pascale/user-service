package domain

import (
	"errors"
	"testing"
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
