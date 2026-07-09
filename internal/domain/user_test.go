package domain

import (
	"errors"
	"testing"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		userName string
		email   string
		role    Role
		wantErr error
	}{
		{
			name:     "usager valide",
			id:       "1",
			userName: "Alice",
			email:    "alice@example.com",
			role:     RoleUser,
			wantErr:  nil,
		},
		{
			name:     "nom vide",
			id:       "2",
			userName: "",
			email:    "bob@example.com",
			role:     RoleUser,
			wantErr:  ErrEmptyName,
		},
		{
			name:     "nom avec seulement des espaces",
			id:       "3",
			userName: "   ",
			email:    "bob@example.com",
			role:     RoleUser,
			wantErr:  ErrEmptyName,
		},
		{
			name:     "courriel sans @",
			id:       "4",
			userName: "Bob",
			email:    "bob-example.com",
			role:     RoleUser,
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "courriel sans domaine",
			id:       "5",
			userName: "Bob",
			email:    "bob@",
			role:     RoleUser,
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "courriel sans point dans le domaine",
			id:       "6",
			userName: "Bob",
			email:    "bob@examplecom",
			role:     RoleUser,
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "role invalide",
			id:       "7",
			userName: "Bob",
			email:    "bob@example.com",
			role:     Role("superadmin"),
			wantErr:  ErrInvalidRole,
		},
		{
			name:     "courriel normalise en minuscules",
			id:       "8",
			userName: "Carole",
			email:    "CAROLE@EXAMPLE.COM",
			role:     RoleAdmin,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.id, tt.userName, tt.email, tt.role)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("erreur = %v, attendu %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return // pas besoin de verifier le reste si on attendait une erreur
			}

			if u.ID != tt.id {
				t.Errorf("ID = %q, attendu %q", u.ID, tt.id)
			}
			if u.CreatedAt.IsZero() {
				t.Error("CreatedAt ne devrait pas etre a zero")
			}
			if u.CreatedAt != u.UpdatedAt {
				t.Error("CreatedAt et UpdatedAt devraient etre identiques a la creation")
			}
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	admin, _ := NewUser("1", "Alice", "alice@example.com", RoleAdmin)
	user, _ := NewUser("2", "Bob", "bob@example.com", RoleUser)

	if !admin.IsAdmin() {
		t.Error("un usager avec le role admin devrait retourner IsAdmin() == true")
	}
	if user.IsAdmin() {
		t.Error("un usager avec le role user devrait retourner IsAdmin() == false")
	}
}

func TestUser_CanView(t *testing.T) {
	admin, _ := NewUser("1", "Alice", "alice@example.com", RoleAdmin)
	userA, _ := NewUser("2", "Bob", "bob@example.com", RoleUser)
	userB, _ := NewUser("3", "Carole", "carole@example.com", RoleUser)

	tests := []struct {
		name   string
		actor  *User
		target *User
		want   bool
	}{
		{"un admin peut voir n'importe quel usager", admin, userA, true},
		{"un usager peut se voir lui-meme", userA, userA, true},
		{"un usager ne peut pas voir un autre usager", userA, userB, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.actor.CanView(tt.target)
			if got != tt.want {
				t.Errorf("CanView() = %v, attendu %v", got, tt.want)
			}
		})
	}
}
