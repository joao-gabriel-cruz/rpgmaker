package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
)

// createAdmin implements `api create-admin --email --name --password`.
func createAdmin(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	email := fs.String("email", "", "admin e-mail")
	name := fs.String("name", "", "display name")
	password := fs.String("password", "", "password (min 8 chars); falls back to ADMIN_PASSWORD")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *password == "" {
		*password = os.Getenv("ADMIN_PASSWORD")
	}
	pool, err := setupDatabase(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	u, err := auth.NewService(auth.NewStore(pool)).CreateUser(ctx, *email, *name, *password, true)
	if errors.Is(err, auth.ErrEmailTaken) {
		return fmt.Errorf("an account with e-mail %q already exists", *email)
	}
	if err != nil {
		return err
	}
	fmt.Printf("admin created: %s <%s>\n", u.DisplayName, u.Email)
	return nil
}
