package domain

import (
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewID(t *testing.T) {
	id := NewID()

	if !uuidV4Pattern.MatchString(id) {
		t.Errorf("NewID() = %q, ne correspond pas au format UUID v4 attendu", id)
	}
}

func TestNewID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := NewID()
		if seen[id] {
			t.Fatalf("collision detectee: %q genere deux fois", id)
		}
		seen[id] = true
	}
}
