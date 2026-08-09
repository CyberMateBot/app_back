package config

import (
	"net/url"
	"os"
	"strings"
	"time"
)

// getenvFirst returns the first non-empty environment variable from keys.
func getenvFirst(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

// normalizeDatabaseURL trims whitespace/quotes and common `psql '...'` wrappers
// that users paste from cloud panel connection snippets.
func normalizeDatabaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)
	if strings.HasPrefix(strings.ToLower(raw), "psql ") {
		raw = strings.TrimSpace(raw[4:])
		raw = strings.Trim(raw, `"'`)
	}
	return raw
}

// encodePostgresURLUserinfo percent-encodes user/password when the panel pasted
// a raw password with reserved characters (e.g. `}`), which makes url.Parse fail
// with "invalid userinfo" and previously caused a silent fallback to localhost.
func encodePostgresURLUserinfo(raw string) (string, bool) {
	schemeIdx := strings.Index(raw, "://")
	if schemeIdx < 0 {
		return "", false
	}
	rest := raw[schemeIdx+3:]
	at := strings.Index(rest, "@")
	if at <= 0 {
		return "", false
	}
	userinfo := rest[:at]
	hostAndRest := rest[at+1:]
	colon := strings.Index(userinfo, ":")
	if colon < 0 {
		return "", false
	}
	user := userinfo[:colon]
	pass := userinfo[colon+1:]
	escaped := url.UserPassword(user, pass).String()
	return raw[:schemeIdx+3] + escaped + "@" + hostAndRest, true
}

func parsePostgresURL(raw string) (ConfigPostgres, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ConfigPostgres{}, false
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "postgres" && scheme != "postgresql" {
		return ConfigPostgres{}, false
	}

	cfg := ConfigPostgres{
		SSLMode: "disable",
	}

	if u.User != nil {
		cfg.User = u.User.Username()
		cfg.Pass, _ = u.User.Password()
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	cfg.Host = host
	cfg.Port = port

	dbName := strings.TrimPrefix(u.Path, "/")
	if i := strings.Index(dbName, "?"); i >= 0 {
		dbName = dbName[:i]
	}
	if dbName != "" {
		cfg.DBName = dbName
	}

	if sslMode := u.Query().Get("sslmode"); sslMode != "" {
		cfg.SSLMode = sslMode
	}

	return cfg, true
}

// postgresFromDatabaseURL parses Railway/Heroku/Timeweb-style DATABASE_URL into ConfigPostgres.
func postgresFromDatabaseURL(raw string) (ConfigPostgres, bool) {
	raw = normalizeDatabaseURL(raw)
	if raw == "" {
		return ConfigPostgres{}, false
	}
	if cfg, ok := parsePostgresURL(raw); ok {
		return cfg, true
	}
	if fixed, ok := encodePostgresURLUserinfo(raw); ok {
		if cfg, ok := parsePostgresURL(fixed); ok {
			return cfg, true
		}
	}
	return ConfigPostgres{}, false
}

func mergePostgresFromEnv(cfg ConfigPostgres) ConfigPostgres {
	// Explicit PG_* always win over DATABASE_URL so a wrong/mangled URL password
	// can be overridden without removing DATABASE_URL entirely.
	if v := getenvFirst("PG_HOST", "PGHOST"); v != "" {
		cfg.Host = v
	}
	if v := getenvFirst("PG_PORT", "PGPORT"); v != "" {
		cfg.Port = v
	}
	if v := getenvFirst("PG_USER", "PGUSER"); v != "" {
		cfg.User = v
	}
	if v := getenvFirst("PG_PASS", "PG_PASSWORD", "PGPASSWORD"); v != "" {
		cfg.Pass = v
	}
	if v := getenvFirst("PG_DBNAME", "PGDATABASE"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("PG_SSLMODE"); v != "" {
		cfg.SSLMode = v
	}
	if cfg.SSLRootCert == "" {
		cfg.SSLRootCert = os.Getenv("PG_SSLROOTCERT")
	}
	return cfg
}

func applyPostgresDefaults(cfg ConfigPostgres) ConfigPostgres {
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == "" {
		cfg.Port = "5432"
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.Pass == "" {
		cfg.Pass = "postgres"
	}
	if cfg.DBName == "" {
		cfg.DBName = "postgres"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.DriverLogLevel == "" {
		cfg.DriverLogLevel = "info"
	}
	if cfg.PoolStatPeriod == 0 {
		cfg.PoolStatPeriod = 30 * time.Second
	}
	if cfg.PoolMaxConns == 0 {
		cfg.PoolMaxConns = 10
	}
	if cfg.PoolMinConns == 0 {
		cfg.PoolMinConns = 1
	}
	if cfg.PoolMaxConnLifeTime == 0 {
		cfg.PoolMaxConnLifeTime = time.Hour
	}
	if cfg.PoolMaxConnIdleTime == 0 {
		cfg.PoolMaxConnIdleTime = 30 * time.Minute
	}
	if cfg.PoolHealthCheckPeriod == 0 {
		cfg.PoolHealthCheckPeriod = time.Minute
	}
	return cfg
}
