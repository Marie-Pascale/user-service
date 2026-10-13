package domain

import (
	"testing"
	"time"
)

func TestNewGroup(t *testing.T) {
	t.Run("creates a valid group", func(t *testing.T) {
		g, err := NewGroup("1", "support-team", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(g.MemberUserIDs) != 0 || len(g.RoleIDs) != 0 {
			t.Error("a new group should be empty (no members, no roles)")
		}
		if g.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
		if g.CreatedAt != g.UpdatedAt {
			t.Error("CreatedAt and UpdatedAt should be identical at creation")
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		_, err := NewGroup("1", "", "")
		if err != ErrEmptyGroupName {
			t.Errorf("error = %v, want %v", err, ErrEmptyGroupName)
		}
	})
}

func TestGroup_Membership(t *testing.T) {
	g, _ := NewGroup("1", "support-team", "")
	originalUpdatedAt := g.UpdatedAt

	time.Sleep(time.Millisecond)
	g.AddMember("user-1")
	if !g.HasMember("user-1") {
		t.Error("the group should have user-1 as a member after AddMember")
	}
	if !g.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should advance after AddMember")
	}

	g.AddMember("user-1")
	if len(g.MemberUserIDs) != 1 {
		t.Errorf("MemberUserIDs = %v, a duplicate member should not be added", g.MemberUserIDs)
	}

	g.RemoveMember("user-1")
	if g.HasMember("user-1") {
		t.Error("the group should no longer have user-1 after RemoveMember")
	}
}

func TestGroup_AssignRole(t *testing.T) {
	g, _ := NewGroup("1", "support-team", "")

	g.AssignRole("role-support-agent")
	if len(g.RoleIDs) != 1 || g.RoleIDs[0] != "role-support-agent" {
		t.Fatalf("RoleIDs = %v, want [\"role-support-agent\"]", g.RoleIDs)
	}

	g.AssignRole("role-support-agent")
	if len(g.RoleIDs) != 1 {
		t.Errorf("RoleIDs = %v, a duplicate role should not be added", g.RoleIDs)
	}
}

func TestGroup_UpdateDetails(t *testing.T) {
	t.Run("updates name and description", func(t *testing.T) {
		g, _ := NewGroup("1", "support-team", "old description")
		g.AddMember("user-1")
		g.AssignRole("role-1")

		err := g.UpdateDetails("support-team-fr", "new description")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.Name != "support-team-fr" {
			t.Errorf("Name = %q, want %q", g.Name, "support-team-fr")
		}
		if len(g.MemberUserIDs) != 1 || len(g.RoleIDs) != 1 {
			t.Error("MemberUserIDs and RoleIDs should not be affected by UpdateDetails")
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		g, _ := NewGroup("1", "support-team", "")

		err := g.UpdateDetails("", "description")
		if err != ErrEmptyGroupName {
			t.Errorf("error = %v, want %v", err, ErrEmptyGroupName)
		}
	})
}
