package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("correct horse", hash) {
		t.Error("expected correct password to verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Error("expected wrong password to fail")
	}
	other, _ := HashPassword("correct horse")
	if other == hash {
		t.Error("expected unique salts")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	for _, h := range []string{"", "plain", "$bcrypt$x$y$z$w", "$argon2id$v=19$m=1,t=1,p=1$!!$!!"} {
		if VerifyPassword("x", h) {
			t.Errorf("malformed hash %q verified", h)
		}
	}
}
