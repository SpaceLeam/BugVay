package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	API        APIConfig
	Postgres   PostgresConfig
	Redis      RedisConfig
	ClickHouse ClickHouseConfig
	Worker     WorkerConfig
	Scanner    ScannerConfig
}

type APIConfig struct {
	Port string
	Host string
	Mode string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ClickHouseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type WorkerConfig struct {
	Concurrency int
	RateLimit   int
	Timeout     int
}

type ScannerConfig struct {
	UserAgent  string
	Timeout    int
	MaxRetries int
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Read .env if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	config := &Config{
		API: APIConfig{
			Port: getEnv("API_PORT", "8080"),
			Host: getEnv("API_HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Database: getEnv("POSTGRES_DB", "bugvay"),
			SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       viper.GetInt("REDIS_DB"),
		},
		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:     getEnv("CLICKHOUSE_PORT", "9000"),
			User:     getEnv("CLICKHOUSE_USER", "default"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
			Database: getEnv("CLICKHOUSE_DB", "bugvay"),
		},
		Worker: WorkerConfig{
			Concurrency: viper.GetInt("WORKER_CONCURRENCY"),
			RateLimit:   viper.GetInt("WORKER_RATE_LIMIT"),
			Timeout:     viper.GetInt("WORKER_TIMEOUT"),
		},
		Scanner: ScannerConfig{
			UserAgent:  getEnv("SCANNER_USER_AGENT", "BUGVay/1.0"),
			Timeout:    viper.GetInt("SCANNER_TIMEOUT"),
			MaxRetries: viper.GetInt("SCANNER_MAX_RETRIES"),
		},
	}

	return config, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
