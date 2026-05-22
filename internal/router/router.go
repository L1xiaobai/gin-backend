package router

import (
	"go-test/internal/api"
	"go-test/internal/global"
	"go-test/internal/middleware"
	"go-test/internal/repository"
	"go-test/internal/service"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	

	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())	
	r.Use(middleware.RateLimit())
	r.Use(middleware.Logger(global.Logger))
	r.Use(middleware.Recovery(global.Logger))

	userRepo := repository.NewUserRepository(global.DB)
	userService := service.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userService)
	healthHandler := api.NewHealthHandler()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/user/register", userHandler.Register)
		v1.POST("/user/update", userHandler.UpdateUser)
		v1.GET("/user/list", userHandler.ListUsers)
		v1.POST("/user/delete", userHandler.DeleteUser)
		v1.GET("/health", healthHandler.HealthCheck)
	}

	return r
}
