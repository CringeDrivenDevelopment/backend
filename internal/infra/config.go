package infra

import (
	"errors"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	// DbUrl - Postgres Database connection string
	// Example - "postgres://username:password@localhost:5432/database_name"
	DbUrl string

	DbHost     string `env:"POSTGRES_HOST" env-default:"localhost"`
	DbPort     string `env:"POSTGRES_PORT" env-default:"5432"`
	DbPassword string `env:"POSTGRES_PASSWORD" env-default:"password"`
	DbUser     string `env:"POSTGRES_USER" env-default:"postgres"`
	DbName     string `env:"POSTGRES_DB" env-default:"muse"`

	JwtSecret string `env:"JWT_SECRET"`
	AppHash   string `env:"APP_HASH"`
	AppId     int    `env:"APP_ID"`
	BotToken  string `env:"BOT_TOKEN"`

	Debug bool `env:"DEBUG" env-default:"false"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	cfg.DbUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)

	if cfg.JwtSecret == "" {
		return nil, errors.New("JWT_SECRET is REQUIRED not to be null")
	}

	if cfg.AppHash == "" {
		return nil, errors.New("APP_HASH is REQUIRED not to be null")
	}

	if cfg.AppId == 0 {
		return nil, errors.New("APP_ID is REQUIRED not to be null")
	}

	if cfg.BotToken == "" {
		return nil, errors.New("BOT_TOKEN is REQUIRED not to be null")
	}

	return &cfg, nil
}
