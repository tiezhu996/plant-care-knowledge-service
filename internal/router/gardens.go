package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerGardenRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc, h *handler.UserGardenHandler, limiter *middleware.RateLimiter) {
	gardens := v1.Group("/gardens", auth)
	gardens.GET("", h.List)
	gardens.POST("", limiter.Limit(), h.Add)
	gardens.PUT("/:id/reminder", h.BindReminder)
	gardens.DELETE("/:id", h.Remove)
}
