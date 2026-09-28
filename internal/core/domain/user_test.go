package domain

import (
	"errors"
	"strings"
	"testing"

	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

func TestNormalizeLogin(t *testing.T) {
	if got := NormalizeLogin("  DaNiL "); got != "danil" {
		t.Fatalf("NormalizeLogin() = %q, want %q", got, "danil")
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr bool
	}{
		{name: "ok", login: "danil_29"},
		{name: "min len", login: "abc"},
		{name: "max len", login: strings.Repeat("a", 32)},
		{name: "too short", login: "ab", wantErr: true},
		{name: "too long", login: strings.Repeat("a", 33), wantErr: true},
		{name: "upper case", login: "Danil", wantErr: true},
		{name: "cyrillic", login: "данил", wantErr: true},
		{name: "space", login: "da nil", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogin(tt.login)
			checkInvalidArgument(t, err, tt.wantErr)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "ok", password: "supersecret"},
		{name: "too short", password: "1234567", wantErr: true},
		{name: "72 bytes", password: strings.Repeat("a", 72)},
		{name: "73 bytes", password: strings.Repeat("a", 73), wantErr: true},
		// 37 кириллических символов = 74 байта: считаем именно байты
		{name: "cyrillic over 72 bytes", password: strings.Repeat("я", 37), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			checkInvalidArgument(t, err, tt.wantErr)
		})
	}
}

func checkInvalidArgument(t *testing.T, err error, wantErr bool) {
	t.Helper()

	if !wantErr {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return
	}

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("error = %v, want ErrInvalidArgument", err)
	}
}
