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
			name:     "valid user",
			id:       "1",
			userName: "Alice",
			email:    "alice@example.com",
			wantErr:  nil,
		},
		{
			name:     "empty name",
			id:       "2",
			userName: "",
			email:    "bob@example.com",
			wantErr:  ErrEmptyName,
		},
		{
			name:     "name containing only whitespace",
			id:       "3",
			userName: "   ",
			email:    "bob@example.com",
			wantErr:  ErrEmptyName,
		},
		{
			name:     "email without @",
			id:       "4",
			userName: "Bob",
			email:    "bob-example.com",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "email without a domain",
			id:       "5",
			userName: "Bob",
			email:    "bob@",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "email without a dot in the domain",
			id:       "6",
			userName: "Bob",
			email:    "bob@examplecom",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "email normalized to lowercase",
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
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if u.ID != tt.id {
				t.Errorf("ID = %q, want %q", u.ID, tt.id)
			}
			if u.Status != StatusActive {
				t.Errorf("Status = %q, want %q (a new user should be active)", u.Status, StatusActive)
			}
			if u.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if u.CreatedAt != u.UpdatedAt {
				t.Error("CreatedAt and UpdatedAt should be identical at creation")
			}
			if u.RoleIDs == nil || u.PermissionIDs == nil {
				t.Error("RoleIDs and PermissionIDs should be initialized (empty slices, not nil)")
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	if !u.IsActive() {
		t.Error("a new user should be active by default")
	}

	u.Suspend()

	if u.IsActive() {
		t.Error("a suspended user should no longer be active")
	}
}

func TestUser_AssignRole(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	u.AssignRole("role-admin")
	if len(u.RoleIDs) != 1 || u.RoleIDs[0] != "role-admin" {
		t.Fatalf("RoleIDs = %v, want [\"role-admin\"]", u.RoleIDs)
	}

	u.AssignRole("role-admin")
	if len(u.RoleIDs) != 1 {
		t.Errorf("RoleIDs = %v, a duplicate role should not be added", u.RoleIDs)
	}
}

func TestUser_AssignPermission(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	u.AssignPermission("perm-orders-read")
	if len(u.PermissionIDs) != 1 || u.PermissionIDs[0] != "perm-orders-read" {
		t.Fatalf("PermissionIDs = %v, want [\"perm-orders-read\"]", u.PermissionIDs)
	}

	u.AssignPermission("perm-orders-read")
	if len(u.PermissionIDs) != 1 {
		t.Errorf("PermissionIDs = %v, a duplicate permission should not be added", u.PermissionIDs)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	t.Run("updates with a valid name and email", func(t *testing.T) {
		u, _ := NewUser("1", "Alice", "alice@example.com")
		u.AssignRole("role-1")

		err := u.UpdateProfile("Alice Dupont", "ALICE.DUPONT@EXAMPLE.COM")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Name != "Alice Dupont" {
			t.Errorf("Name = %q, want %q", u.Name, "Alice Dupont")
		}
		if u.Email != "alice.dupont@example.com" {
			t.Errorf("Email = %q, want the lowercase version", u.Email)
		}
		if len(u.RoleIDs) != 1 {
			t.Error("RoleIDs should not be affected by UpdateProfile")
		}
	})

	t.Run("rejects an empty name without modifying the user", func(t *testing.T) {
		u, _ := NewUser("1", "Alice", "alice@example.com")

		err := u.UpdateProfile("", "nouveau@example.com")
		if err != ErrEmptyName {
			t.Errorf("error = %v, want %v", err, ErrEmptyName)
		}
		if u.Name != "Alice" {
			t.Error("Name should not have changed after validation failed")
		}
	})

	t.Run("rejects an invalid email address", func(t *testing.T) {
		u, _ := NewUser("1", "Alice", "alice@example.com")

		err := u.UpdateProfile("Alice", "not-an-email")
		if err != ErrInvalidEmail {
			t.Errorf("error = %v, want %v", err, ErrInvalidEmail)
		}
	})
}

func TestUser_RecordLogin(t *testing.T) {
	u, _ := NewUser("1", "Alice", "alice@example.com")

	if u.LastLoginAt != nil {
		t.Error("LastLoginAt should be nil before any login")
	}

	originalUpdatedAt := u.UpdatedAt
	time.Sleep(time.Millisecond)
	u.RecordLogin()

	if u.LastLoginAt == nil {
		t.Fatal("LastLoginAt should no longer be nil after RecordLogin")
	}
	if !u.LastLoginAt.After(originalUpdatedAt) {
		t.Error("LastLoginAt should be later than the user creation time")
	}
	if !u.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should also advance after RecordLogin")
	}
}
