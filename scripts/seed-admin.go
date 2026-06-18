//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()

	email := strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL")))
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		fmt.Println("set ADMIN_EMAIL and ADMIN_PASSWORD in .env")
		os.Exit(1)
	}

	host := getenv("PG_HOST", "localhost")
	port := getenv("PG_PORT", "5432")
	user := getenv("PG_USER", "postgres")
	pass := getenv("PG_PASSWORD", getenv("PG_PASS", "postgres"))
	db := getenv("PG_DBNAME", "myapp_db")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Println("db connect:", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("hash:", err)
		os.Exit(1)
	}

	tag, err := conn.Exec(ctx, `
INSERT INTO admins(email, password_hash, role)
VALUES($1, $2, 'admin')
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash`,
		email, string(hash))
	if err != nil {
		fmt.Println("upsert:", err)
		os.Exit(1)
	}
	fmt.Printf("admin upserted: %s (rows %d)\n", email, tag.RowsAffected())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
