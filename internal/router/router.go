package router

import (
	"github.com/MukizuL/diploma-1/internal/controller"
	mw "github.com/MukizuL/diploma-1/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func newRouter(c *controller.Controller, mw *mw.MiddlewareService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.POST("/api/user/register", c.Register)
	router.POST("/api/user/login", c.Login)

	withAuth := router.Group("/api/user").Use(mw.Authorization())

	withAuth.POST("/orders", c.PostOrders)
	withAuth.GET("/orders", c.GetOrders)
	withAuth.GET("/balance", c.GetBalance)
	withAuth.POST("/balance/withdraw", c.Withdraw)
	withAuth.GET("/withdrawals", c.GetWithdrawals)

	return router
}

func Provide() fx.Option {
	return fx.Provide(newRouter)
}
