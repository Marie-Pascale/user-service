package domain

import (
	"testing"
	"time"
)

func TestNewRole(t *testing.T) {
	t.Run("cree un role valide", func(t *testing.T) {
		r, err := NewRole("1", "support-agent", "agent de support")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if len(r.PermissionIDs) != 0 {
			t.Errorf("un nouveau role ne devrait avoir aucune permission, a: %v", r.PermissionIDs)
		}
		if r.CreatedAt.IsZero() {
			t.Error("CreatedAt ne devrait pas etre a zero")
		}
		if r.CreatedAt != r.UpdatedAt {
			t.Error("CreatedAt et UpdatedAt devraient etre identiques a la creation")
		}
	})

	t.Run("rejette un nom vide", func(t *testing.T) {
		_, err := NewRole("1", "", "description")
		if err != ErrEmptyRoleName {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyRoleName)
		}
	})
}

func TestRole_GrantAndRevokePermission(t *testing.T) {
	r, _ := NewRole("1", "support-agent", "")
	originalUpdatedAt := r.UpdatedAt

	time.Sleep(time.Millisecond)
	r.GrantPermission("perm-orders-read")
	if !r.HasPermission("perm-orders-read") {
		t.Error("le role devrait avoir la permission apres GrantPermission")
	}
	if !r.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt devrait avancer apres GrantPermission")
	}

	// Accorder deux fois ne doit pas creer de doublon.
	r.GrantPermission("perm-orders-read")
	if len(r.PermissionIDs) != 1 {
		t.Errorf("PermissionIDs = %v, une permission dupliquee ne devrait pas etre ajoutee", r.PermissionIDs)
	}

	r.RevokePermission("perm-orders-read")
	if r.HasPermission("perm-orders-read") {
		t.Error("le role ne devrait plus avoir la permission apres RevokePermission")
	}
}

func TestRole_UpdateDetails(t *testing.T) {
	t.Run("met a jour nom et description", func(t *testing.T) {
		r, _ := NewRole("1", "support-agent", "ancienne description")
		r.GrantPermission("perm-1") // pour verifier que ca survit a la mise a jour

		err := r.UpdateDetails("support-agent-senior", "nouvelle description")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if r.Name != "support-agent-senior" {
			t.Errorf("Name = %q, attendu %q", r.Name, "support-agent-senior")
		}
		if r.Description != "nouvelle description" {
			t.Errorf("Description = %q, attendu %q", r.Description, "nouvelle description")
		}
		if len(r.PermissionIDs) != 1 {
			t.Error("PermissionIDs ne devrait pas etre affecte par UpdateDetails")
		}
	})

	t.Run("rejette un nom vide", func(t *testing.T) {
		r, _ := NewRole("1", "support-agent", "")

		err := r.UpdateDetails("", "description")
		if err != ErrEmptyRoleName {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyRoleName)
		}
		if r.Name != "support-agent" {
			t.Error("Name ne devrait pas avoir change apres une validation echouee")
		}
	})
}
