package common

import "github.com/caarlos0/env"

type Config struct {
}

func NewConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, err
	}

	return c, nil
}
