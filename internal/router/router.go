package router

import (
	"go-test/internal/api"
	"go-test/internal/global"
	"go-test/internal/middleware"
	"go-test/internal/repository"
	"go-test/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	

	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:8080", "http://127.0.0.1:8080"}, // 允许访问的前端域名
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    }))
	r.Use(middleware.RequestID())	
	r.Use(middleware.Logger(global.Logger))
	r.Use(middleware.Recovery(global.Logger))

	userRepo := repository.NewUserRepository(global.DB)
	userService := service.NewUserService(userRepo)
	userHandler := api.NewUserHandler(userService)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/user/register", userHandler.Register)
		v1.POST("/user/update", userHandler.UpdateUser)
		v1.GET("/user/list", userHandler.ListUsers)
		v1.POST("/user/delete", userHandler.DeleteUser)
	}

	return r
}
