package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env"
)

type configENV struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
	KafkaAdress string `env:"KAFKA_ADDRESS"`
}

func GetConfig() (*configENV, error) {
	var eCfg configENV

	flag.StringVar(&eCfg.KafkaAdress, "a",
		"localhost:9092",
		"address where work kafka")
	flag.StringVar(&eCfg.DatabaseDSN, "d",
		"postgres://user:password@localhost:5432/local?sslmode=disable",
		"address database connection")
	flag.Parse()

	err := env.Parse(&eCfg)
	if err != nil {
		return nil, fmt.Errorf("failed parsing environment variables: %w", err)
	}

	flag.Parse()

	return &eCfg, nil
}
