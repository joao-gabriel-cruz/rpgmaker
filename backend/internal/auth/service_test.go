package auth

import (
	"context"
	"errors"
	"testing"
)

func seeded(t *testing.T) (*Service, *fakeRepo) {
	t.Helper()
	repo := newFakeRepo()
	svc := NewService(repo)
	if _, err := svc.CreateUser(context.Background(), "Ana@Example.com", "Ana", "secret-pass", true); err != nil {
		t.Fatal(err)
	}
	return svc, repo
}

func TestLoginSucceedsCaseInsensitive(t *testing.T) {
	svc, repo := seeded(t)
	_, token, _, err := svc.Login(context.Background(), "ana@example.com", "secret-pass")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.sessions[token]; !ok {
		t.Error("session not stored")
	}
}

func TestLoginWrongCredentialsAreIndistinguishable(t *testing.T) {
	svc, _ := seeded(t)
	_, _, _, wrongPass := svc.Login(context.Background(), "ana@example.com", "nope")
	_, _, _, wrongEmail := svc.Login(context.Background(), "x@example.com", "secret-pass")
	if !errors.Is(wrongPass, ErrInvalidCredentials) || !errors.Is(wrongEmail, ErrInvalidCredentials) {
		t.Errorf("got %v / %v", wrongPass, wrongEmail)
	}
}

func TestCreateUserValidatesAndRejectsDuplicates(t *testing.T) {
	svc, _ := seeded(t)
	if _, err := svc.CreateUser(context.Background(), "b@x.com", "B", "short", false); err == nil {
		t.Error("expected short password error")
	}
	if _, err := svc.CreateUser(context.Background(), "ANA@example.com", "A2", "secret-pass", false); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("got %v", err)
	}
}

func TestResetPasswordRevokesSessionsAndNewPasswordWorks(t *testing.T) {
	svc, repo := seeded(t)
	ctx := context.Background()
	_, token, _, _ := svc.Login(ctx, "ana@example.com", "secret-pass")
	temp, err := svc.ResetPassword(ctx, "id-Ana@Example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.sessions[token]; ok {
		t.Error("old session should be revoked")
	}
	if _, _, _, err := svc.Login(ctx, "ana@example.com", "secret-pass"); err == nil {
		t.Error("old password should stop working")
	}
	if _, _, _, err := svc.Login(ctx, "ana@example.com", temp); err != nil {
		t.Errorf("temp password rejected: %v", err)
	}
	if _, err := svc.ResetPassword(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v", err)
	}
}
