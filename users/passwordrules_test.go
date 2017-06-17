package users

import (
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	err := readCommonPasswords(strings.NewReader("buzzword\n"))
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"ok", "swordfish", false},
		{"ok-long", "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", false},
		{"short", "1234", true},
		{"singleChar", "11111111", true},
		{"repetitive", "12121212", true},
		{"toolong", "abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", true},
		{"common", "buzzword", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePassword(tt.password); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
