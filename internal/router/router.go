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
	r.Use(middleware.RequestID())	
	r.Use(middleware.Logger(global.Logger))

	userRepo := repository.NewUserRepository(global.DB)
	userService := service.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userService)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/user/register", userHandler.Register)
	}

	return r
}
