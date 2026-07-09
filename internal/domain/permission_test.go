package domain

import "testing"

func TestNewPermission(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		key       string
		wantErr   error
	}{
		{"permission valide", "service-1", "orders:read", nil},
		{"serviceID vide", "", "orders:read", ErrEmptyServiceID},
		{"cle vide", "service-1", "", ErrEmptyPermissionKey},
		{"cle avec seulement des espaces", "service-1", "   ", ErrEmptyPermissionKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPermission("1", tt.serviceID, tt.key, "description")

			if err != tt.wantErr {
				t.Fatalf("erreur = %v, attendu %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && p.Key != tt.key {
				t.Errorf("Key = %q, attendu %q", p.Key, tt.key)
			}
		})
	}
}
