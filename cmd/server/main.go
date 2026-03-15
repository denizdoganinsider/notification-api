package main

import (
	"net/http"
	"notification-api/config"
	"notification-api/internal/controller"
	"notification-api/internal/repository"
	"notification-api/internal/service"

	"github.com/labstack/echo/v4"

	notificationMiddleware "notification-api/internal/middleware"
)

func main() {
	db := config.NewDatabase()
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	e.POST("/register", userController.Register)
	e.POST("/login", userController.Login)
	e.GET("/me", userController.Me, notificationMiddleware.JWTMiddleware)

	e.Logger.Fatal(e.Start(":8080"))
}
