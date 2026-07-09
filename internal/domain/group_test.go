package domain

import "testing"

func TestNewGroup(t *testing.T) {
	t.Run("cree un groupe valide", func(t *testing.T) {
		g, err := NewGroup("1", "equipe-support", "")
		if err != nil {
			t.Fatalf("erreur inattendue: %v", err)
		}
		if len(g.MemberUserIDs) != 0 || len(g.RoleIDs) != 0 {
			t.Error("un nouveau groupe devrait etre vide (aucun membre, aucun role)")
		}
	})

	t.Run("rejette un nom vide", func(t *testing.T) {
		_, err := NewGroup("1", "", "")
		if err != ErrEmptyGroupName {
			t.Errorf("erreur = %v, attendu %v", err, ErrEmptyGroupName)
		}
	})
}

func TestGroup_Membership(t *testing.T) {
	g, _ := NewGroup("1", "equipe-support", "")

	g.AddMember("user-1")
	if !g.HasMember("user-1") {
		t.Error("le groupe devrait avoir user-1 comme membre apres AddMember")
	}

	// Ajouter deux fois ne doit pas creer de doublon.
	g.AddMember("user-1")
	if len(g.MemberUserIDs) != 1 {
		t.Errorf("MemberUserIDs = %v, un membre duplique ne devrait pas etre ajoute", g.MemberUserIDs)
	}

	g.RemoveMember("user-1")
	if g.HasMember("user-1") {
		t.Error("le groupe ne devrait plus avoir user-1 apres RemoveMember")
	}
}

func TestGroup_AssignRole(t *testing.T) {
	g, _ := NewGroup("1", "equipe-support", "")

	g.AssignRole("role-support-agent")
	if len(g.RoleIDs) != 1 || g.RoleIDs[0] != "role-support-agent" {
		t.Fatalf("RoleIDs = %v, attendu [\"role-support-agent\"]", g.RoleIDs)
	}

	g.AssignRole("role-support-agent")
	if len(g.RoleIDs) != 1 {
		t.Errorf("RoleIDs = %v, un role duplique ne devrait pas etre ajoute", g.RoleIDs)
	}
}
