package domain

import (
	"testing"
	"time"
)

func TestNewRole(t *testing.T) {
	t.Run("creates a valid role", func(t *testing.T) {
		r, err := NewRole("1", "support-agent", "agent de support")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(r.PermissionIDs) != 0 {
			t.Errorf("a new role should have no permissions; got: %v", r.PermissionIDs)
		}
		if r.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
		if r.CreatedAt != r.UpdatedAt {
			t.Error("CreatedAt and UpdatedAt should be identical at creation")
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		_, err := NewRole("1", "", "description")
		if err != ErrEmptyRoleName {
			t.Errorf("error = %v, want %v", err, ErrEmptyRoleName)
		}
	})
}

func TestRole_GrantAndRevokePermission(t *testing.T) {
	r, _ := NewRole("1", "support-agent", "")
	originalUpdatedAt := r.UpdatedAt

	time.Sleep(time.Millisecond)
	r.GrantPermission("perm-orders-read")
	if !r.HasPermission("perm-orders-read") {
		t.Error("the role should have the permission after GrantPermission")
	}
	if !r.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should advance after GrantPermission")
	}

	r.GrantPermission("perm-orders-read")
	if len(r.PermissionIDs) != 1 {
		t.Errorf("PermissionIDs = %v, a duplicate permission should not be added", r.PermissionIDs)
	}

	r.RevokePermission("perm-orders-read")
	if r.HasPermission("perm-orders-read") {
		t.Error("the role should no longer have the permission after RevokePermission")
	}
}

func TestRole_UpdateDetails(t *testing.T) {
	t.Run("updates name and description", func(t *testing.T) {
		r, _ := NewRole("1", "support-agent", "old description")
		r.GrantPermission("perm-1")

		err := r.UpdateDetails("support-agent-senior", "new description")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Name != "support-agent-senior" {
			t.Errorf("Name = %q, want %q", r.Name, "support-agent-senior")
		}
		if r.Description != "new description" {
			t.Errorf("Description = %q, want %q", r.Description, "new description")
		}
		if len(r.PermissionIDs) != 1 {
			t.Error("PermissionIDs should not be affected by UpdateDetails")
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		r, _ := NewRole("1", "support-agent", "")

		err := r.UpdateDetails("", "description")
		if err != ErrEmptyRoleName {
			t.Errorf("error = %v, want %v", err, ErrEmptyRoleName)
		}
		if r.Name != "support-agent" {
			t.Error("Name should not have changed after validation failed")
		}
	})
}
