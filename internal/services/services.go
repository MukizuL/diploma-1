package services

import (
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/MukizuL/diploma-1/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Services struct {
	storage storage.Repo
	logger  *zap.Logger
	key     []byte
}

func newServices(storage storage.Repo, logger *zap.Logger, cfg *config.Config) *Services {
	return &Services{
		storage: storage,
		logger:  logger,
		key:     []byte(cfg.Key),
	}
}

func Provide() fx.Option {
	return fx.Provide(newServices)
}
