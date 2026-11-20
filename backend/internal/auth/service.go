package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ErrInvalidCredentials is returned for any failed login, without saying which field was wrong.
var ErrInvalidCredentials = errors.New("invalid credentials")

// SessionTTL is how long a login lasts.
const SessionTTL = 30 * 24 * time.Hour

const tempPasswordAlphabet = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// Repository is the persistence the Service needs.
type Repository interface {
	CreateUser(ctx context.Context, email, hash, name string, admin bool) (User, error)
	UserByEmail(ctx context.Context, email string) (User, error)
	UserByID(ctx context.Context, id string) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	SetPassword(ctx context.Context, id, hash string) error
	CreateSession(ctx context.Context, id, userID string, expires time.Time) error
	UserBySession(ctx context.Context, id string) (User, error)
	DeleteSession(ctx context.Context, id string) error
}

// Service holds the account and session rules.
type Service struct {
	repo      Repository
	now       func() time.Time
	dummyHash string
}

// NewService builds a Service.
func NewService(repo Repository) *Service {
	dummy, _ := HashPassword("unused-password")
	return &Service{repo: repo, now: time.Now, dummyHash: dummy}
}

// Login checks credentials and opens a session, returning its token and expiry.
func (s *Service) Login(ctx context.Context, email, password string) (User, string, time.Time, error) {
	u, err := s.repo.UserByEmail(ctx, strings.TrimSpace(email))
	switch {
	case errors.Is(err, ErrNotFound):
		VerifyPassword(password, s.dummyHash) // keep timing similar for unknown e-mails
		return User{}, "", time.Time{}, ErrInvalidCredentials
	case err != nil:
		return User{}, "", time.Time{}, err
	case !VerifyPassword(password, u.PasswordHash):
		return User{}, "", time.Time{}, ErrInvalidCredentials
	}
	token, err := randomToken()
	if err != nil {
		return User{}, "", time.Time{}, err
	}
	expires := s.now().Add(SessionTTL)
	if err := s.repo.CreateSession(ctx, token, u.ID, expires); err != nil {
		return User{}, "", time.Time{}, err
	}
	return u, token, expires, nil
}

// Authenticate resolves a session token to its user.
func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	return s.repo.UserBySession(ctx, token)
}

// Logout ends a session.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteSession(ctx, token)
}

// CreateUser registers an account with a hashed password.
func (s *Service) CreateUser(ctx context.Context, email, name, password string, admin bool) (User, error) {
	email, name = strings.TrimSpace(email), strings.TrimSpace(name)
	if email == "" || name == "" || len(password) < 8 {
		return User{}, errors.New("email, name and a password of at least 8 characters are required")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	return s.repo.CreateUser(ctx, email, hash, name, admin)
}

// ListUsers returns every account.
func (s *Service) ListUsers(ctx context.Context) ([]User, error) { return s.repo.ListUsers(ctx) }

// ResetPassword sets a random temporary password, revokes sessions and returns the password once.
func (s *Service) ResetPassword(ctx context.Context, id string) (string, error) {
	if _, err := s.repo.UserByID(ctx, id); err != nil {
		return "", err
	}
	temp, err := tempPassword(12)
	if err != nil {
		return "", err
	}
	hash, err := HashPassword(temp)
	if err != nil {
		return "", err
	}
	return temp, s.repo.SetPassword(ctx, id, hash)
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func tempPassword(n int) (string, error) {
	out := make([]byte, n)
	limit := big.NewInt(int64(len(tempPasswordAlphabet)))
	for i := range out {
		idx, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate password: %w", err)
		}
		out[i] = tempPasswordAlphabet[idx.Int64()]
	}
	return string(out), nil
}
