package controller

import (
	"errors"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

func (c *Controller) Register(ctx *gin.Context) {
	var data dto.AuthFormIn
	err := ctx.BindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &gin.H{
			"Error":   http.StatusText(http.StatusBadRequest),
			"Message": err.Error(),
		})
		return
	}

	token, err := c.service.CreateUser(ctx.Request.Context(), data.Login, data.Password)
	if err != nil {
		if errors.Is(err, errs.ErrConflictLogin) {
			ctx.JSON(http.StatusConflict, &gin.H{
				"Error": err.Error(),
			})
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Register"), zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.SetCookie("Access-token", token, 3600, "/", c.domain, false, true)

	ctx.JSON(http.StatusOK, &gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}

func (c *Controller) Login(ctx *gin.Context) {
	var data dto.AuthFormIn
	err := ctx.BindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &gin.H{
			"Error":   http.StatusText(http.StatusBadRequest),
			"Message": err.Error(),
		})
		return
	}

	token, err := c.service.LoginUser(ctx.Request.Context(), data.Login, data.Password)
	if err != nil {
		if errors.Is(err, errs.ErrNotAuthorized) {
			ctx.JSON(http.StatusUnauthorized, &gin.H{
				"Error":   http.StatusText(http.StatusUnauthorized),
				"Message": "Access token is invalid",
			})
			return
		}

		if errors.Is(err, errs.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, &gin.H{
				"Error":   http.StatusText(http.StatusUnauthorized),
				"Message": "Login or password is incorrect",
			})
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Login"), zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.SetCookie("Access-token", token, 3600, "/", c.domain, false, true)

	ctx.JSON(http.StatusOK, &gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}

func (c *Controller) PostOrders(ctx *gin.Context) {
	if ctx.GetHeader("Content-type") != "text/plain" {
		ctx.JSON(http.StatusBadRequest, &gin.H{
			"Error":   http.StatusText(http.StatusBadRequest),
			"Message": "Only accepts text/plain",
		})
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		c.logger.Error("Error in handler", zap.String("handler", "PostOrders"), zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	if len(data) > 18 {
		ctx.JSON(http.StatusUnprocessableEntity, &gin.H{
			"Error":   http.StatusText(http.StatusUnprocessableEntity),
			"Message": "Invalid order ID",
		})
		return
	}

	orderID, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, &gin.H{
			"Error":   http.StatusText(http.StatusUnprocessableEntity),
			"Message": "OrderOut ID must be a number",
		})
		return
	}

	userID := ctx.MustGet("userID").(string)

	err = c.service.PostOrder(ctx.Request.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, errs.ErrWrongOrderFormat) {
			ctx.JSON(http.StatusUnprocessableEntity, &gin.H{
				"Error":   http.StatusText(http.StatusUnprocessableEntity),
				"Message": "Invalid order ID",
			})
			return
		}

		if errors.Is(err, errs.ErrConflictOrder) {
			ctx.JSON(http.StatusConflict, &gin.H{
				"Error":   http.StatusText(http.StatusConflict),
				"Message": err.Error(),
			})
			return
		}

		if errors.Is(err, errs.ErrDuplicateOrder) {
			ctx.JSON(http.StatusOK, &gin.H{
				"Result": "This order is already uploaded",
			})
			return
		}
	}

	ctx.JSON(http.StatusAccepted, &gin.H{
		"Result": http.StatusText(http.StatusAccepted),
	})
}

func (c *Controller) GetOrders(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	orders, err := c.service.GetOrders(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			ctx.Status(http.StatusNoContent)
			return
		}

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

func (c *Controller) GetBalance(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)
	balance, err := c.service.GetBalance(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			ctx.JSON(http.StatusUnauthorized, &gin.H{
				"Error":   http.StatusText(http.StatusUnauthorized),
				"Message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	c.logger.Info("Balance in handler", zap.Any("balance", balance))

	ctx.JSON(http.StatusOK, balance)
}

func (c *Controller) Withdraw(ctx *gin.Context) {
	var data dto.OrderIn
	err := ctx.BindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &gin.H{
			"Error":   http.StatusText(http.StatusBadRequest),
			"Message": err.Error(),
		})
		return
	}

	orderID, err := strconv.ParseInt(data.OrderID, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, &gin.H{
			"Error":   http.StatusText(http.StatusUnprocessableEntity),
			"Message": "OrderOut ID must be a number",
		})
		return
	}

	userID := ctx.MustGet("userID").(string)

	err = c.service.PostOrderWithWithdrawal(ctx, userID, orderID, data.Sum)
	if err != nil {
		if errors.Is(err, errs.ErrWrongOrderFormat) {
			ctx.JSON(http.StatusUnprocessableEntity, &gin.H{
				"Error":   http.StatusText(http.StatusUnprocessableEntity),
				"Message": "Invalid order ID",
			})
			return
		}

		if errors.Is(err, errs.ErrConflictOrder) {
			ctx.JSON(http.StatusConflict, &gin.H{
				"Error":   http.StatusText(http.StatusConflict),
				"Message": err.Error(),
			})
			return
		}

		if errors.Is(err, errs.ErrInsufficientBalance) {
			ctx.JSON(http.StatusPaymentRequired, &gin.H{
				"Error":   http.StatusText(http.StatusPaymentRequired),
				"Message": err.Error(),
			})
			return
		}

		if errors.Is(err, errs.ErrDuplicateOrder) {
			ctx.JSON(http.StatusOK, &gin.H{
				"Result": "This order is already uploaded",
			})
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Withdraw"), zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}
}

func (c *Controller) GetWithdrawals(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	orders, err := c.service.GetWithdrawals(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrWithdrawalNotFound) {
			ctx.Status(http.StatusNoContent)
			return
		}

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}
