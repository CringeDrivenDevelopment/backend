package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Settings struct {
	// DbUrl - Postgres Database connection string
	// Example - "postgres://username:password@localhost:5432/database_name"
	DbUrl          string
	JwtSecret      string `env:"JWT_SECRET"`
	BotToken       string `env:"BOT_TOKEN"`
	YoutubeUrl     string `env:"YOUTUBE_URL" env-default:"https://yt.lxft.tech"`
	YoutubeToken   string `env:"YOUTUBE_TOKEN"`
	MinioBucket    string `env:"MINIO_BUCKET" env-default:"music"`
	MinioHost      string `env:"MINIO_HOST" env-default:"localhost:9000"`
	MinioSecretKey string `env:"MINIO_SECRET_KEY" env-default:"minioadmin"`
	MinioAccessKey string `env:"MINIO_ACCESS_KEY" env-default:"minioadmin"`

	DbHost     string `env:"POSTGRES_HOST" env-default:"localhost"`
	DbPort     string `env:"POSTGRES_PORT" env-default:"5432"`
	DbPassword string `env:"POSTGRES_PASSWORD" env-default:"password"`
	DbUser     string `env:"POSTGRES_USER" env-default:"user"`
	DbName     string `env:"POSTGRES_DB" env-default:"db"`
}

func New() (*Settings, error) {
	var cfg Settings
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	cfg.DbUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)

	if cfg.JwtSecret == "" {
		return nil, errors.New("JWT_SECRET is REQUIRED not to be null")
	}

	if cfg.YoutubeToken == "" {
		return nil, errors.New("YOUTUBE_TOKEN is REQUIRED not to be null")
	}

	if cfg.BotToken == "" {
		return nil, errors.New("BOT_TOKEN is REQUIRED not to be null")
	}

	return &cfg, nil
}
