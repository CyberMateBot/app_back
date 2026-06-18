package config

import (
	"os"
	"strconv"
	"strings"
	"time"

)

type Config struct {
	Postgres ConfigPostgres
	App      ConfigApp
	JWT      ConfigJWT
	Redis    ConfigRedis
	CORS     ConfigCORS
	Server   ConfigServer
}

type ConfigApp struct {
	HTTPPort       int
	GRPCPort       int
	Environment    string
	Debug          bool
	LogLevel       string
	SwaggerEnabled bool
	APITitle       string
	APIVersion     string

	// Telegram deep links for Mini App UI (Support button, referral links).
	SupportTelegramInviteURL    string
	TelegramBotUsername         string
	TelegramReferralParamPrefix string // e.g. ref_ → startapp=ref_{telegram_id}
}

type ConfigServer struct {
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type ConfigPostgres struct {
	Host           string
	Port           string
	User           string
	Pass           string
	DBName         string
	SSLMode        string
	SSLRootCert    string
	Debug          bool
	DriverLogLevel string

	PoolStatPeriod        time.Duration
	PoolMaxConns          int64
	PoolMinConns          int64
	PoolMaxConnLifeTime   time.Duration
	PoolMaxConnIdleTime   time.Duration
	PoolHealthCheckPeriod time.Duration
}

type ConfigJWT struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type ConfigRedis struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ConfigCORS struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}

func getenvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func trimEnvQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return strings.TrimSpace(s[1 : len(s)-1])
		}
	}
	return s
}

func getenvStringSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := trimEnvQuotes(p); s != "" {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return def
}

// CORSAllowsAll reports whether any configured origin is a wildcard.
func CORSAllowsAll(origins []string) bool {
	for _, o := range origins {
		if strings.TrimSpace(o) == "*" {
			return true
		}
	}
	return false
}

func LoadConfig() Config {
	LoadDotEnv()

	return Config{
		Postgres: LoadPostgresConfig(),
		App:      LoadAppConfig(),
		JWT:      LoadJWTConfig(),
		Redis:    LoadRedisConfig(),
		CORS:     LoadCORSConfig(),
		Server:   LoadServerConfig(),
	}
}

func LoadAppConfig() ConfigApp {
	return ConfigApp{
		HTTPPort:       getenvInt("APP_HTTP_PORT", 8090),
		GRPCPort:       getenvInt("APP_GRPC_PORT", 8091),
		Environment:    getenv("ENVIRONMENT", "development"),
		Debug:          getenvBool("DEBUG", false),
		LogLevel:       getenv("LOG_LEVEL", "info"),
		SwaggerEnabled: getenvBool("SWAGGER_ENABLED", true),
		APITitle:       getenv("API_TITLE", "Your API"),
		APIVersion:     getenv("API_VERSION", "1.0.0"),

		SupportTelegramInviteURL:    getenv("TELEGRAM_SUPPORT_INVITE_URL", "https://t.me/+jXI2qDR9Y-xkZTI6"),
		TelegramBotUsername:         getenv("TELEGRAM_BOT_USERNAME", "CyberMate_bot"),
		TelegramReferralParamPrefix: getenv("TELEGRAM_REFERRAL_PARAM_PREFIX", "ref_"),
	}
}

func LoadServerConfig() ConfigServer {
	return ConfigServer{
		Host:         getenv("SERVER_HOST", "localhost"),
		ReadTimeout:  getenvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getenvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getenvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
	}
}

func LoadPostgresConfig() ConfigPostgres {
	var cfg ConfigPostgres

	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		if parsed, ok := postgresFromDatabaseURL(raw); ok {
			cfg = parsed
		}
	}

	cfg = mergePostgresFromEnv(cfg)
	cfg = applyPostgresDefaults(cfg)

	if v := os.Getenv("PG_SSLMODE"); v != "" {
		cfg.SSLMode = v
	}

	cfg.Debug = getenvBool("PG_DEBUG", cfg.Debug)
	if v := os.Getenv("PG_DRIVER_LOG_LEVEL"); v != "" {
		cfg.DriverLogLevel = v
	}
	if v := os.Getenv("PG_SSLROOTCERT"); v != "" {
		cfg.SSLRootCert = v
	}

	cfg.PoolStatPeriod = getenvDuration("PG_POOL_STAT_PERIOD", cfg.PoolStatPeriod)
	cfg.PoolMaxConns = getenvInt64("PG_POOL_MAX_CONNS", cfg.PoolMaxConns)
	cfg.PoolMinConns = getenvInt64("PG_POOL_MIN_CONNS", cfg.PoolMinConns)
	cfg.PoolMaxConnLifeTime = getenvDuration("PG_POOL_MAX_CONN_LIFETIME", cfg.PoolMaxConnLifeTime)
	cfg.PoolMaxConnIdleTime = getenvDuration("PG_POOL_MAX_CONN_IDLE_TIME", cfg.PoolMaxConnIdleTime)
	cfg.PoolHealthCheckPeriod = getenvDuration("PG_POOL_HEALTH_CHECK_PERIOD", cfg.PoolHealthCheckPeriod)

	return cfg
}

func LoadJWTConfig() ConfigJWT {
	return ConfigJWT{
		Secret:          getenv("JWT_SECRET", "your-super-secret-jwt-key-change-this"),
		AccessTokenTTL:  getenvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getenvDuration("JWT_REFRESH_TOKEN_TTL", 168*time.Hour), // 7 days
	}
}

func LoadRedisConfig() ConfigRedis {
	return ConfigRedis{
		Host:     getenv("REDIS_HOST", "localhost"),
		Port:     getenv("REDIS_PORT", "6379"),
		Password: getenv("REDIS_PASSWORD", ""),
		DB:       getenvInt("REDIS_DB", 0),
	}
}

func LoadCORSConfig() ConfigCORS {
	defaultOrigins := []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173"}
	if strings.EqualFold(getenv("ENVIRONMENT", ""), "production") {
		defaultOrigins = []string{"*"}
	}
	return ConfigCORS{
		AllowedOrigins: getenvStringSlice("CORS_ALLOWED_ORIGINS", defaultOrigins),
		AllowedMethods: getenvStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		AllowedHeaders: getenvStringSlice("CORS_ALLOWED_HEADERS", []string{
			"Content-Type", "Authorization", "Accept", "X-Requested-With",
		}),
	}
}
