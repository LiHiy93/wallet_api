package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `env:"ENV" env-default:"local"`
	HTTPPort    string        `env:"HTTP_PORT" env-default:"8080"`
	DatabaseURL string        `env:"DATABASE_URL" env-required:"true"`
	JWTSecret   string        `env:"JWT_SECRET" env-required:"true"`
	JWTTTL      time.Duration `env:"JWT_TTL" env-default:"24h"`
	Migrations  string        `env:"MIGRATIONS_PATH" env-default:"file://migrations"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	if path != "" {
		if err := cleanenv.ReadConfig(path, cfg); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		return cfg, nil
	}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("read env config: %w", err)
	}
	return cfg, nil
}
