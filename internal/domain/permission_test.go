package domain

import "testing"

func TestNewPermission(t *testing.T) {
	tests := []struct {
		name      string
		serviceID string
		key       string
		wantErr   error
	}{
		{"valid permission", "service-1", "orders:read", nil},
		{"empty serviceID", "", "orders:read", ErrEmptyServiceID},
		{"empty key", "service-1", "", ErrEmptyPermissionKey},
		{"key containing only whitespace", "service-1", "   ", ErrEmptyPermissionKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPermission("1", tt.serviceID, tt.key, "description")

			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if p.Key != tt.key {
				t.Errorf("Key = %q, want %q", p.Key, tt.key)
			}
			if p.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if p.CreatedAt != p.UpdatedAt {
				t.Error("CreatedAt and UpdatedAt should be identical at creation")
			}
		})
	}
}
