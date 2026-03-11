// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBDsn         string
	RedisAddr     string
	RedisPassword string
	EncryptKey    string
	CastDir       string
	JWTSecret     string
	Port          int
}

var C Config

func Load() error {
	C = Config{
		DBDsn:         requireEnv("SSHIVE_DB_DSN"),
		RedisAddr:     requireEnv("SSHIVE_REDIS_ADDR"),
		RedisPassword: os.Getenv("SSHIVE_REDIS_PASSWORD"),
		EncryptKey:    requireEnv("SSHIVE_ENCRYPT_KEY"),
		CastDir:       getEnvOrDefault("SSHIVE_CAST_DIR", "/data/sshive/cast"),
		JWTSecret:     requireEnv("SSHIVE_JWT_SECRET"),
	}
	portStr := getEnvOrDefault("SSHIVE_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid SSHIVE_PORT: %w", err)
	}
	C.Port = port
	return nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
