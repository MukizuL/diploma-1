package services

import (
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/MukizuL/diploma-1/internal/storage"
	"github.com/MukizuL/diploma-1/internal/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Services struct {
	storage storage.Repo
	logger  *zap.Logger
	worker  *worker.Worker
	key     []byte
}

func newServices(storage storage.Repo, logger *zap.Logger, cfg *config.Config, worker *worker.Worker) *Services {
	return &Services{
		storage: storage,
		logger:  logger,
		worker:  worker,
		key:     []byte(cfg.Key),
	}
}

func Provide() fx.Option {
	return fx.Provide(newServices)
}
