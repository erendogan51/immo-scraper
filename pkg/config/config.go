package config

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

var config *Config

func GetConfig(ctx context.Context, locations ...string) (Config, error) {
	if config != nil {
		return *config, nil
	}

	cfg, err := newConfig(ctx, locations...)
	if err != nil {
		return Config{}, err
	}

	config = &cfg

	return *config, nil
}

func newConfig(ctx context.Context, locations ...string) (Config, error) {
	config := &Config{}

	if err := godotenv.Load(locations...); err != nil {
		log.Warn().Msg("could not load .env file, will use the existing ones")
	}

	if err := envconfig.Process(ctx, config); err != nil {
		return Config{}, fmt.Errorf("processing environment config: %w", err)
	}

	return *config, nil
}
