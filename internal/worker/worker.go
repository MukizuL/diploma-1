package worker

import (
	"context"
	"encoding/json"
	"github.com/MukizuL/diploma-1/internal/config"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"resty.dev/v3"
	"strconv"
	"time"
)

type Worker struct {
	in      chan int64
	done    chan struct{}
	logger  *zap.Logger
	storage storage.Repo
	c       *resty.Client
	cfg     *config.Config
}

func (w *Worker) Push(orderID int64) error {
	select {
	case <-w.done:
		return errs.ErrWorkerIsDone
	default:
		w.in <- orderID
		return nil
	}
}

func (w *Worker) Run() {
	for {
		select {
		case <-w.done:
			return
		case order := <-w.in:
			resp, err := w.c.R().Get("/api/orders/" + strconv.FormatInt(order, 10))
			if err != nil {
				w.logger.Error("Failed to query accrual system", zap.Int64("orderID", order), zap.Error(err))
				w.Push(order)
				continue
			}

			var data dto.AccrualResp
			if resp.StatusCode() == http.StatusOK {
				err = json.NewDecoder(resp.Body).Decode(&data)
				if err != nil {
					w.logger.Error("Failed to unmarshal accrual system response", zap.Int64("orderID", order), zap.Error(err))
					w.Push(order)
					continue
				}
				resp.Body.Close()
			}

			switch resp.StatusCode() {
			case http.StatusOK:
				switch data.Status {
				case "PROCESSED":
					w.logger.Info("Processed by accrual system", zap.Any("order", data))
					err = w.storage.UpdateOrderWithAccrual(context.TODO(), order, "PROCESSED", data.Accrual)
					if err != nil {
						w.logger.Error("Failed to update order", zap.Int64("orderID", order), zap.Error(err))
					}
				case "REGISTERED", "PROCESSING":
					w.logger.Info("Still processing by accrual system", zap.Int64("orderID", order))
					err = w.storage.UpdateOrder(context.TODO(), order, "PROCESSING")
					if err != nil {
						w.logger.Error("Failed to update order", zap.Int64("orderID", order), zap.Error(err))
					}
					go func() {
						<-time.After(time.Second)
						w.Push(order)
					}()
				case "INVALID":
					w.logger.Info("Invalidated by accrual system", zap.Int64("orderID", order))
					err = w.storage.UpdateOrder(context.TODO(), order, "INVALID")
					if err != nil {
						w.logger.Error("Failed to update order", zap.Int64("orderID", order), zap.Error(err))
					}
				}
			case http.StatusNoContent:
				w.logger.Info("Not registered by accrual system", zap.Int64("orderID", order))
			case http.StatusTooManyRequests:
				w.logger.Info("Too many requests to accrual system", zap.Int64("orderID", order))
				<-time.After(10 * time.Second)
				w.Push(order)
			}
		}
	}
}

func (w *Worker) Shutdown() {
	close(w.done)
	w.c.Close()
}

func newWorker(lc fx.Lifecycle, logger *zap.Logger, cfg *config.Config, storage storage.Repo) *Worker {
	c := resty.New().
		SetBaseURL("http://"+cfg.AccrualSystem).
		SetHeader("Content-Length", "0")

	w := &Worker{
		in:      make(chan int64, 100),
		done:    make(chan struct{}),
		logger:  logger,
		storage: storage,
		c:       c,
		cfg:     cfg,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go w.Run()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			w.Shutdown()

			return nil
		},
	})

	return w
}

func Provide() fx.Option {
	return fx.Provide(newWorker)
}
