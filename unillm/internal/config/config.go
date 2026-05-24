package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port         string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	BcryptCost   int
	Environment   string // dev, staging, prod
	CORSOrigins   []string
	MaxBodyBytes  int64
	EncryptionKey string // hex-encoded 32-byte AES-256 key for provider key encryption
}

func Load() Config {
	bcryptCost := 12
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			bcryptCost = n
		}
	}

	env := getEnv("ENVIRONMENT", "dev")
	jwtSecret := getEnv("JWT_SECRET", "dev_jwt_secret_change_in_prod")

	// In production, enforce non-default JWT secret
	if env == "prod" && jwtSecret == "dev_jwt_secret_change_in_prod" {
		log.Fatal("FATAL: JWT_SECRET must be set to a secure value in production")
	}

	corsOrigins := []string{"*"}
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		corsOrigins = strings.Split(v, ",")
		for i := range corsOrigins {
			corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
		}
	}

	maxBody := int64(10 * 1024 * 1024) // 10MB default
	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxBody = n
		}
	}

	encKey := getEnv("ENCRYPTION_KEY", "")
	if env == "prod" && encKey == "" {
		log.Fatal("FATAL: ENCRYPTION_KEY must be set in production for provider key encryption")
	}

	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://unillm:unillm_dev_2026@localhost:5432/unillm?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:    jwtSecret,
		BcryptCost:   bcryptCost,
		Environment:  env,
		CORSOrigins:  corsOrigins,
		MaxBodyBytes:  maxBody,
		EncryptionKey: encKey,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
