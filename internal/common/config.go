package common

import "github.com/caarlos0/env"

type Config struct {
	FirebaseProjectID string `env:"FIREBASE_PROJECT_ID" envDefault:""`
}

func NewConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	return c, nil
}
