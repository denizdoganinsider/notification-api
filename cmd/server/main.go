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

	notificationRepo := repository.NewNotificationRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	cacheService := service.NewCacheService()
	webhookDeliveryService := service.NewWebhookDeliveryService(webhookRepo)
	eventService := service.NewEventService(cacheService, webhookDeliveryService)

	notificationService := service.NewNotificationService(notificationRepo, eventService)
	notificationController := controller.NewNotificationController(notificationService)

	webhookService := service.NewWebhookService(webhookRepo)
	webhookController := controller.NewWebhookController(webhookService)

	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	e.POST("/register", userController.Register)
	e.POST("/login", userController.Login)

	auth := e.Group("")
	auth.Use(notificationMiddleware.JWTMiddleware)

	auth.GET("/me", userController.Me)

	auth.POST("/notifications", notificationController.Create)
	auth.GET("/notifications", notificationController.List)
	auth.GET("/notifications/:id", notificationController.GetByID)

	auth.POST("/webhooks", webhookController.Create)
	auth.GET("/webhooks", webhookController.List)
	auth.DELETE("/webhooks/:id", webhookController.Delete)

	e.Logger.Fatal(e.Start(":8080"))
}
