package controller

import (
	"github.com/MukizuL/diploma-1/internal/services"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Controller struct {
	service *services.Services
	logger  *zap.Logger
}

func newController(logger *zap.Logger, service *services.Services) *Controller {
	return &Controller{
		service: service,
		logger:  logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newController)
}
