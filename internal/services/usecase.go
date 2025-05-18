package services

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/helpers"
)

func (s *Services) PostOrder(ctx context.Context, userID string, orderID int64) error {
	if !helpers.ValidLuhn(orderID) {
		return errs.ErrWrongOrderFormat
	}

	user, err := s.storage.GetOrderByID(ctx, orderID)
	if err != nil && !errors.Is(err, errs.ErrOrderNotFound) {
		return err
	}

	if user != "" {
		if userID != user {
			return errs.ErrConflictOrder
		} else {
			return errs.ErrDuplicateOrder
		}
	}

	err = s.storage.CreateNewOrder(ctx, userID, orderID)
	if err != nil {
		return err
	}

	return nil
}
