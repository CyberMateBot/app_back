//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	host := env("PG_HOST", "localhost")
	port := env("PG_PORT", "5432")
	user := env("PG_USER", "postgres")
	pass := firstNonEmpty(os.Getenv("PG_PASSWORD"), os.Getenv("PG_PASS"), os.Getenv("PGPASSWORD"))
	db := env("PG_DBNAME", "postgres")
	ssl := env("PG_SSLMODE", "require")

	// Keyword DSN avoids URL-encoding issues with special chars in passwords (e.g. '}').
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, db, ssl,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fail("connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fail("ping: %v", err)
	}
	fmt.Println("connected OK")

	root, err := os.Getwd()
	if err != nil {
		fail("cwd: %v", err)
	}
	// Allow running from repo root or scripts/
	migDir := filepath.Join(root, "internal", "migrations")
	if _, err := os.Stat(migDir); err != nil {
		migDir = filepath.Join(root, "..", "internal", "migrations")
	}

	files := []string{
		"V20250805000000__users.sql",
		"V20251103000100__cybermate_core.sql",
		"V20251103000200__admin_resources.sql",
		"V20250525000000__profile_ui_theme.sql",
		"V20260527000000__prompt_history.sql",
		"V20260528000100__web_auth_and_prompts.sql",
		"V20260607000100__profile_is_active.sql",
		"V20260608000100__token_transactions.sql",
		"V20260608000200__wallets_profile_unique.sql",
		"V20260620000100__admin_panel_extras.sql",
		"V20260616000100__home_widgets.sql",
		"V20260621000100__nullable_token_tx_admin.sql",
		"V20260622000100__user_subscriptions.sql",
		"V20260724000100__rebalance_billing_catalog.sql",
	}

	for _, name := range files {
		path := filepath.Join(migDir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			fail("read %s: %v", name, err)
		}
		fmt.Printf("Applying %s ...\n", name)
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			fail("apply %s: %v", name, err)
		}
		fmt.Printf("OK %s\n", name)
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('profiles','admins','wallets')`).Scan(&n); err != nil {
		fail("verify: %v", err)
	}
	fmt.Printf("Done. Found %d/3 core tables (profiles, admins, wallets).\n", n)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
