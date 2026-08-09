package config

import "testing"

func TestPostgresFromDatabaseURL(t *testing.T) {
	raw := "postgresql://user:secret@db.example.com:6543/mydb?sslmode=require"
	cfg, ok := postgresFromDatabaseURL(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if cfg.Host != "db.example.com" || cfg.Port != "6543" || cfg.User != "user" || cfg.Pass != "secret" || cfg.DBName != "mydb" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestPostgresFromDatabaseURL_RawSpecialCharPassword(t *testing.T) {
	// Timeweb often pastes passwords with `}` unencoded — url.Parse rejects that as invalid userinfo.
	raw := "postgresql://gen_user:ebmIWt4CCv}Vsg@24879e6f20791375a1bc3a29.twc1.net:5432/default_db?sslmode=require"
	cfg, ok := postgresFromDatabaseURL(raw)
	if !ok {
		t.Fatal("expected ok for raw special-char password")
	}
	if cfg.User != "gen_user" || cfg.Pass != "ebmIWt4CCv}Vsg" || cfg.Host != "24879e6f20791375a1bc3a29.twc1.net" || cfg.DBName != "default_db" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestPostgresFromDatabaseURL_EncodedSpecialCharPassword(t *testing.T) {
	raw := "postgresql://gen_user:ebmIWt4CCv%7DVsg@24879e6f20791375a1bc3a29.twc1.net:5432/default_db?sslmode=require"
	cfg, ok := postgresFromDatabaseURL(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	if cfg.Pass != "ebmIWt4CCv}Vsg" {
		t.Fatalf("pass=%q, want decoded }", cfg.Pass)
	}
}

func TestLoadPostgresConfig_RailwayEnvNames(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PG_HOST", "")
	t.Setenv("PGHOST", "postgres.railway.internal")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGUSER", "railway")
	t.Setenv("PGPASSWORD", "pw")
	t.Setenv("PGDATABASE", "railway")
	t.Setenv("PG_SSLMODE", "require")

	cfg := LoadPostgresConfig()
	if cfg.Host != "postgres.railway.internal" || cfg.User != "railway" || cfg.Pass != "pw" || cfg.DBName != "railway" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLoadPostgresConfig_DatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@host:5432/db?sslmode=require")
	t.Setenv("PG_HOST", "")
	t.Setenv("PGHOST", "")
	t.Setenv("PG_PASSWORD", "")
	t.Setenv("PGPASSWORD", "")
	t.Setenv("PG_USER", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PG_DBNAME", "")
	t.Setenv("PGDATABASE", "")

	cfg := LoadPostgresConfig()
	if cfg.Host != "host" || cfg.DBName != "db" || cfg.User != "u" || cfg.Pass != "p" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLoadPostgresConfig_PGPasswordOverridesURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:wrong@host:5432/db?sslmode=require")
	t.Setenv("PG_PASSWORD", "correct")
	t.Setenv("PGPASSWORD", "")
	t.Setenv("PG_PASS", "")

	cfg := LoadPostgresConfig()
	if cfg.Pass != "correct" || cfg.Host != "host" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}
