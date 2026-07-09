package domain

import "testing"

func TestNewRole(t *testing.T) {
	t.Run("cree un role valide", func(t *testing.T) {
		r, err := NewRole("1", "support-agent", "agent de support")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if len(r.PermissionIDs) != 0 {
			t.Errorf("un nouveau role ne devrait avoir aucune permission, a: %v", r.PermissionIDs)
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

	r.GrantPermission("perm-orders-read")
	if !r.HasPermission("perm-orders-read") {
		t.Error("le role devrait avoir la permission apres GrantPermission")
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
