package storage

import (
	"context"
	"github.com/MukizuL/diploma-1/internal/models"
	"github.com/MukizuL/diploma-1/internal/storage/pg"
	"go.uber.org/fx"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go -package=mockstorage

type Repo interface {
	CreateNewUser(ctx context.Context, login, passwordHash string) (string, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, string, error)
	GetOrderByID(ctx context.Context, orderID int64) (string, error)
	CreateNewOrder(ctx context.Context, userID string, orderID int64) error
}

func newRepo(storage *pg.Storage) Repo {
	return storage
}

func Provide() fx.Option {
	return fx.Provide(newRepo)
}
