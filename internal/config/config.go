package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      App
	Debug    Debug
	Postgres Postgres
	Redis    Redis
	JWT      JWT
	Cache    Cache
	Log      Log
}

type App struct {
	Addr string
	Env  string
}

type Debug struct {
	Enabled    bool
	PathPrefix string
	Token      string
}

type Postgres struct {
	DSN string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

type JWT struct {
	Secret    string
	Issuer    string
	AccessTTL time.Duration
}

type Cache struct {
	LastMessagesLimit int
	LastMessagesTTL   time.Duration
	UnreadTTL         time.Duration
}

type Log struct {
	LevelStr string
}

func (l Log) Level() slog.Level {
	switch l.LevelStr {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func MustLoad() Config {
	return Config{
		App: App{
			Addr: getenv("APP_ADDR", ":8080"),
			Env:  getenv("APP_ENV", "development"),
		},
		Debug: Debug{
			Enabled:    getenvBool("DEBUG_ENABLED", true),
			PathPrefix: getenv("DEBUG_PATH_PREFIX", "/debug"),
			Token:      getenv("DEBUG_TOKEN", ""),
		},
		Postgres: Postgres{
			DSN: getenv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/messaging?sslmode=disable"),
		},
		Redis: Redis{
			Addr:     getenv("REDIS_ADDR", "localhost:6379"),
			Password: getenv("REDIS_PASSWORD", ""),
			DB:       getenvInt("REDIS_DB", 0),
		},
		JWT: JWT{
			Secret:    mustGetenv("JWT_SECRET"),
			Issuer:    getenv("JWT_ISSUER", "messaging-api"),
			AccessTTL: getenvDuration("JWT_ACCESS_TTL", 24*time.Hour),
		},
		Cache: Cache{
			LastMessagesLimit: getenvInt("CACHE_LAST_MESSAGES_LIMIT", 100),
			LastMessagesTTL:   getenvDuration("CACHE_LAST_MESSAGES_TTL", 15*time.Minute),
			UnreadTTL:         getenvDuration("CACHE_UNREAD_TTL", 20*time.Minute),
		},
		Log: Log{LevelStr: getenv("LOG_LEVEL", "info")},
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getenvDuration(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getenvBool(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
