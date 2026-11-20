package auth

import (
	"context"
	"strings"
	"time"
)

type fakeRepo struct {
	users    map[string]User
	sessions map[string]string
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[string]User{}, sessions: map[string]string{}}
}

func (f *fakeRepo) CreateUser(_ context.Context, email, hash, name string, admin bool) (User, error) {
	for _, u := range f.users {
		if strings.EqualFold(u.Email, email) {
			return User{}, ErrEmailTaken
		}
	}
	u := User{ID: "id-" + email, Email: email, PasswordHash: hash, DisplayName: name, IsAdmin: admin}
	f.users[u.ID] = u
	return u, nil
}

func (f *fakeRepo) UserByEmail(_ context.Context, email string) (User, error) {
	for _, u := range f.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}

func (f *fakeRepo) UserByID(_ context.Context, id string) (User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

func (f *fakeRepo) ListUsers(context.Context) ([]User, error) {
	out := []User{}
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

func (f *fakeRepo) SetPassword(_ context.Context, id, hash string) error {
	u := f.users[id]
	u.PasswordHash = hash
	f.users[id] = u
	for t, uid := range f.sessions {
		if uid == id {
			delete(f.sessions, t)
		}
	}
	return nil
}

func (f *fakeRepo) CreateSession(_ context.Context, id, userID string, _ time.Time) error {
	f.sessions[id] = userID
	return nil
}

func (f *fakeRepo) UserBySession(_ context.Context, id string) (User, error) {
	uid, ok := f.sessions[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return f.users[uid], nil
}

func (f *fakeRepo) DeleteSession(_ context.Context, id string) error {
	delete(f.sessions, id)
	return nil
}
