package storage

import (
	"context"
	"github.com/MukizuL/diploma-1/internal/models"
	"github.com/MukizuL/diploma-1/internal/storage/pg"
	"github.com/greatcloak/decimal"
	"go.uber.org/fx"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go -package=mockstorage

type Repo interface {
	CreateNewUser(ctx context.Context, login, passwordHash string) (string, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByOrderID(ctx context.Context, orderID int64) (string, error)
	CreateNewOrder(ctx context.Context, userID string, orderID int64) error
	CreateNewOrderWithWithdrawal(ctx context.Context, userID string, orderID int64, sum decimal.Decimal) error
	GetOrdersByUser(ctx context.Context, userID string) ([]models.Order, error)
	GetWithdrawalsByUser(ctx context.Context, userID string) ([]models.Withdrawal, error)
	GetBalance(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error)
	UpdateOrder(ctx context.Context, orderID int64, status string) error
	UpdateOrderWithAccrual(ctx context.Context, orderID int64, status string, accrual decimal.Decimal) error
}

func newRepo(storage *pg.Storage) Repo {
	return storage
}

func Provide() fx.Option {
	return fx.Provide(newRepo)
}
