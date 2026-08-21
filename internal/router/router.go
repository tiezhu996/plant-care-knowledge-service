package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
	"github.com/gbplantwiki/gbplantwiki/internal/dto"
	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
	"github.com/gbplantwiki/gbplantwiki/internal/repository"
	"github.com/gbplantwiki/gbplantwiki/internal/service"
	"gorm.io/gorm"
)

// Setup builds the gin engine with all routes and middleware.
func Setup(cfg *config.Config, db *gorm.DB, logger *slog.Logger) *gin.Engine {
	// repositories
	userRepo := repository.NewUserRepository(db)
	plantRepo := repository.NewPlantSpeciesRepository(db)
	articleRepo := repository.NewCareArticleRepository(db)
	pestRepo := repository.NewDiseasePestRepository(db)
	reminderRepo := repository.NewCareReminderRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	gardenRepo := repository.NewUserGardenRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	answerRepo := repository.NewAnswerRepository(db)

	// services
	userService := service.NewUserService(userRepo, logger, cfg)
	plantService := service.NewPlantSpeciesService(plantRepo, logger)
	articleService := service.NewCareArticleService(articleRepo, logger)
	pestService := service.NewDiseasePestService(pestRepo, logger)
	reminderService := service.NewCareReminderService(reminderRepo, logger)
	favoriteService := service.NewFavoriteService(favoriteRepo, logger)
	gardenService := service.NewUserGardenService(gardenRepo, logger)
	questionService := service.NewQuestionService(questionRepo, answerRepo, userService, logger)
	answerService := service.NewAnswerService(db, answerRepo, questionRepo, logger)

	// handlers
	userHandler := handler.NewUserHandler(userService, logger)
	plantHandler := handler.NewPlantSpeciesHandler(plantService, logger)
	articleHandler := handler.NewCareArticleHandler(articleService, logger)
	pestHandler := handler.NewDiseasePestHandler(pestService, logger)
	reminderHandler := handler.NewCareReminderHandler(reminderService, logger)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService, logger)
	gardenHandler := handler.NewUserGardenHandler(gardenService, logger)
	questionHandler := handler.NewQuestionHandler(questionService, logger)
	answerHandler := handler.NewAnswerHandler(answerService, logger)
	uploadHandler := handler.NewUploadHandler(cfg, logger)
	homeHandler := handler.NewHomeHandler(plantService, articleService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.ErrorHandler(logger))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, dto.OK(gin.H{"status": "ok"})) })

	limiter := middleware.NewRateLimiter(cfg.RateLimitReq, cfg.RateLimitWin)
	auth := middleware.AuthRequired(cfg, userRepo)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/home/overview", homeHandler.Overview)
		registerUserRoutes(v1, auth, userHandler, limiter)
		registerPlantRoutes(v1, auth, plantHandler, limiter)
		registerArticleRoutes(v1, auth, articleHandler, limiter)
		registerPestRoutes(v1, auth, pestHandler, limiter)
		registerReminderRoutes(v1, auth, reminderHandler, limiter)
		registerFavoriteRoutes(v1, auth, favoriteHandler, limiter)
		registerGardenRoutes(v1, auth, gardenHandler, limiter)
		registerQuestionRoutes(v1, auth, questionHandler, answerHandler, limiter)
		v1.POST("/uploads", auth, limiter.Limit(), uploadHandler.Upload)
		v1.PUT("/answers/:id/like", auth, answerHandler.Like)
	}
	return r
}
