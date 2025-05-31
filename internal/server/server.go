package server

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func newHTTPServer(lc fx.Lifecycle, cfg *config.Config, router *gin.Engine, logger *zap.Logger) *http.Server {
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: router.Handler(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting HTTP server", zap.String("addr", cfg.Addr))

			var err error

			go func() {
				err = srv.ListenAndServe()
			}()

			time.Sleep(100 * time.Millisecond)

			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func Provide() fx.Option {
	return fx.Provide(newHTTPServer)
}
