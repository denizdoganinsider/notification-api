package main

import (
	"net/http"
	"notification-api/config"
	"notification-api/internal/controller"
	"notification-api/internal/repository"
	"notification-api/internal/service"
	"time"

	"github.com/labstack/echo/v4"

	notificationMiddleware "notification-api/internal/middleware"
)

func main() {
	db := config.NewDatabase()
	defer db.Close()

	redisClient := config.NewRedisClient()
	defer redisClient.Close()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	notificationRepo := repository.NewNotificationRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	cacheService := service.NewCacheService(redisClient)
	webhookDeliveryService := service.NewWebhookDeliveryService(webhookRepo)
	eventService := service.NewEventService(cacheService, webhookDeliveryService)

	notificationService := service.NewNotificationService(notificationRepo, eventService, cacheService)
	notificationController := controller.NewNotificationController(notificationService)

	webhookService := service.NewWebhookService(webhookRepo)
	webhookController := controller.NewWebhookController(webhookService)

	adminController := controller.NewAdminController(userService, notificationService)

	e := echo.New()

	e.Use(notificationMiddleware.RateLimiterMiddleware(redisClient, 100, 1*time.Minute))

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

	admin := auth.Group("/admin")
	admin.Use(notificationMiddleware.AdminMiddleware)

	admin.GET("/users", adminController.ListUsers)
	admin.GET("/notifications", adminController.ListNotifications)

	e.Logger.Fatal(e.Start(":8080"))
}
