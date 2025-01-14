package common

import "github.com/caarlos0/env"

type Config struct {
	FirebaseProjectID string `env:"FIREBASE_PROJECT_ID" envDefault:""`
	DatabaseFQDN      string `env:"DATABASE_FQDN"`
	Workers           int    `env:"WORKERS" envDefault:"5"`
	WorkerPoolBuffer  int    `env:"WORKER_POOL_BUFFER" envDefault:"10"`
}

func NewConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	return c, nil
}
