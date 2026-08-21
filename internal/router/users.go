package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerUserRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc, h *handler.UserHandler, limiter *middleware.RateLimiter) {
	users := v1.Group("/users")
	users.POST("/register", limiter.Limit(), h.Register)
	users.POST("/login", limiter.Limit(), h.Login)
	me := users.Group("/me", auth)
	me.GET("", h.GetProfile)
	me.PUT("", h.UpdateProfile)
}
