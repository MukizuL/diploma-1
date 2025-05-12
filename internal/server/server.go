package server

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
	"time"
)

func newHTTPServer(lc fx.Lifecycle, cfg *config.Config, router *gin.Engine, logger *zap.Logger) *http.Server {
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: router.Handler(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := Migrate(cfg.DSN)
			if err != nil {
				return err
			}

			logger.Info("Starting HTTP server", zap.String("addr", cfg.Addr))

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

func Migrate(DSN string) error {
	_, err := filepath.Abs("./migrations")
	if err != nil {
		return err
	}

	m, err := migrate.New("file://migrations", DSN+"?sslmode=disable")
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
