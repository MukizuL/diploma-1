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
	var data dto.AuthForm
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

	ctx.SetCookie("Access-token", token, 3600, "/", "", true, true)

	ctx.JSON(http.StatusOK, &gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}

func (c *Controller) Login(ctx *gin.Context) {
	var data dto.AuthForm
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

	ctx.SetCookie("Access-token", token, 3600, "/", "", true, true)

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

	if len(data) != 11 {
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
			"Message": "Order ID must be a number",
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

	ctx.JSON(http.StatusCreated, &gin.H{
		"Result": http.StatusText(http.StatusCreated),
	})
}

func (c *Controller) GetOrders(ctx *gin.Context) {
	userID := ctx.MustGet("userID").(string)

	orders, err := c.service.GetOrders(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			ctx.JSON(http.StatusNoContent, nil)
			return
		}

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}
