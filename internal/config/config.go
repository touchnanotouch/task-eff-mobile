package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host    string
	Port    string
	Timeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var (
	once sync.Once
	cfg  Config
)

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

func Validate(cfg Config) error {
	if cfg.Server.Host == "" || cfg.Server.Port == "" {
		return fmt.Errorf("server host or port is empty")
	}

	if cfg.Database.Host == "" || cfg.Database.Port == "" {
		return fmt.Errorf("database host or port is empty")
	}

	if cfg.Database.User == "" || cfg.Database.Password == "" || cfg.Database.DBName == "" {
		return fmt.Errorf("database user, password and dbname are required")
	}

	if cfg.Database.SSLMode == "" {
		return fmt.Errorf("database sslmode is empty")
	}

	return nil
}

func Load() *Config {
	once.Do(func() {
		cfg = Config{
			Server: ServerConfig{
				Host:    getEnv("SERVER_HOST", "localhost"),
				Port:    getEnv("SERVER_PORT", "8080"),
				Timeout: time.Duration(getEnvInt("SERVER_TIMEOUT", 30)) * time.Second,
			},

			Database: DatabaseConfig{
				Host:     getEnv("DATABASE_HOST", "localhost"),
				Port:     getEnv("DATABASE_PORT", "5432"),
				User:     getEnv("DATABASE_USER", ""),
				Password: getEnv("DATABASE_PASSWORD", ""),
				DBName:   getEnv("DATABASE_NAME", ""),
				SSLMode:  getEnv("DATABASE_SSL_MODE", "disable"),
			},
		}

		if err := Validate(cfg); err != nil {
			log.Fatalf("invalid config: %v", err)
		}
	})

	return &cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("invalid int value for %s: %s", key, value)
	}

	return intValue
}
