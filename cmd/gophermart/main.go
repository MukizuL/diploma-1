package main

import (
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/MukizuL/diploma-1/internal/controller"
	mw "github.com/MukizuL/diploma-1/internal/middleware"
	"github.com/MukizuL/diploma-1/internal/router"
	"github.com/MukizuL/diploma-1/internal/server"
	"github.com/MukizuL/diploma-1/internal/services"
	"github.com/MukizuL/diploma-1/internal/storage"
	"github.com/MukizuL/diploma-1/internal/storage/pg"
	"github.com/MukizuL/diploma-1/internal/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	fx.New(createApp(), fx.Invoke(func(*http.Server) {}, func(*worker.Worker) {})).Run()
}

func createApp() fx.Option {
	return fx.Options(
		config.Provide(),
		fx.Provide(zap.NewDevelopment),

		controller.Provide(),
		router.Provide(),
		server.Provide(),
		services.Provide(),
		mw.Provide(),

		pg.Provide(),
		storage.Provide(),

		worker.Provide(),
	)
}
