package auth_service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

type fakeUsersRepository struct {
	users map[string]domain.User
}

func newFakeUsersRepository() *fakeUsersRepository {
	return &fakeUsersRepository{users: make(map[string]domain.User)}
}

func (r *fakeUsersRepository) CreateUser(_ context.Context, user domain.User) (domain.User, error) {
	if _, ok := r.users[user.Login]; ok {
		return domain.User{}, fmt.Errorf("login='%s': %w", user.Login, core_errors.ErrConflict)
	}

	user.ID = int64(len(r.users) + 1)
	r.users[user.Login] = user

	return user, nil
}

func (r *fakeUsersRepository) GetUserByLogin(_ context.Context, login string) (domain.User, error) {
	user, ok := r.users[login]
	if !ok {
		return domain.User{}, fmt.Errorf("login='%s': %w", login, core_errors.ErrNotFound)
	}

	return user, nil
}

type fakeTokenManager struct{}

func (fakeTokenManager) NewToken(userID int64) (string, error) {
	return fmt.Sprintf("token-%d", userID), nil
}

func TestRegister(t *testing.T) {
	repo := newFakeUsersRepository()
	service := NewAuthService(repo, fakeTokenManager{})
	ctx := context.Background()

	user, token, err := service.Register(ctx, "  Danil ", "supersecret")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if user.Login != "danil" {
		t.Fatalf("login = %q, want normalized %q", user.Login, "danil")
	}

	if user.PasswordHash == "" || user.PasswordHash == "supersecret" {
		t.Fatalf("password must be stored as hash, got %q", user.PasswordHash)
	}

	if token != "token-1" {
		t.Fatalf("token = %q, want %q", token, "token-1")
	}

	_, _, err = service.Register(ctx, "DANIL", "anotherpass")
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("duplicate login error = %v, want ErrConflict", err)
	}

	_, _, err = service.Register(ctx, "bob", "short")
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("short password error = %v, want ErrInvalidArgument", err)
	}
}

func TestLogin(t *testing.T) {
	repo := newFakeUsersRepository()
	service := NewAuthService(repo, fakeTokenManager{})
	ctx := context.Background()

	if _, _, err := service.Register(ctx, "danil", "supersecret"); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, _, err := service.Login(ctx, "Danil", "supersecret"); err != nil {
		t.Fatalf("Login with correct password: %v", err)
	}

	_, _, errWrongPassword := service.Login(ctx, "danil", "wrongpass1")
	_, _, errUnknownLogin := service.Login(ctx, "nobody", "wrongpass1")

	for name, err := range map[string]error{
		"wrong password": errWrongPassword,
		"unknown login":  errUnknownLogin,
	} {
		if !errors.Is(err, core_errors.ErrUnauthenticated) {
			t.Fatalf("%s: error = %v, want ErrUnauthenticated", name, err)
		}
	}

	if errWrongPassword.Error() != errUnknownLogin.Error() {
		t.Fatalf("errors must be identical, got %q and %q", errWrongPassword, errUnknownLogin)
	}
}
