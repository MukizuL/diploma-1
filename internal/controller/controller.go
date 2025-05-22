package controller

import (
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/MukizuL/diploma-1/internal/services"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Controller struct {
	domain  string
	service *services.Services
	logger  *zap.Logger
}

func newController(logger *zap.Logger, service *services.Services, cfg *config.Config) *Controller {
	return &Controller{
		domain:  cfg.Addr,
		service: service,
		logger:  logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newController)
}
