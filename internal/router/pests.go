package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerPestRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc, h *handler.DiseasePestHandler, limiter *middleware.RateLimiter) {
	pests := v1.Group("/pests")
	pests.GET("", h.List)
	pests.GET("/:id", h.Get)
	admin := pests.Group("", auth, middleware.RequireRole("admin"))
	admin.POST("", limiter.Limit(), h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}
