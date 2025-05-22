package server

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

//go:embed "migrations/*.sql"
var embedMigrations embed.FS

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
	db, err := sql.Open("pgx", DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
