package config

import (
	"errors"
	"flag"
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"
	"log"
	"strconv"
	"strings"
)

var ErrMalformedFlags = errors.New("error parsing flags")
var ErrMalformedAddr = errors.New("address of wrong format")

var ErrEmptyDSN = errors.New("dsn cannot be empty")
var ErrEmptyAccrualSystem = errors.New("accrual system address cannot be empty")

type Config struct {
	Addr          string `env:"RUN_ADDRESS"`
	DSN           string `env:"DATABASE_URI"`
	AccrualSystem string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Key           string `env:"SECRET_KEY"`
}

func newConfig() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal("Error parsing env variables")
	}

	if cfg.Addr == "" {
		flag.StringVar(&cfg.Addr, "a", "0.0.0.0:8080", "Адрес и порт запуска сервиса")
	}

	if cfg.DSN == "" {
		flag.StringVar(&cfg.DSN, "d", "", "Адрес подключения к базе данных")
	}

	if cfg.AccrualSystem == "" {
		flag.StringVar(&cfg.AccrualSystem, "r", "", "Адрес системы расчёта начислений")
	}

	flag.Parse()

	err = checkParams(cfg)
	if err != nil {
		flag.Usage()
		//log.Fatal(err)
		return nil, err
	}

	return cfg, nil
}

func checkParams(cfg *Config) error {
	if cfg.Addr != "" {
		addr := strings.Split(cfg.Addr, ":")
		if len(addr) != 2 {
			return ErrMalformedAddr
		}

		_, err := strconv.Atoi(addr[1])
		if err != nil {
			return ErrMalformedAddr
		}
	}

	if cfg.DSN == "" {
		return ErrEmptyDSN
	}

	if cfg.AccrualSystem == "" {
		return ErrEmptyAccrualSystem
	}

	return nil
}

func Provide() fx.Option {
	return fx.Provide(newConfig)
}
