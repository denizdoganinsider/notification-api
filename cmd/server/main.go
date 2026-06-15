package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"notification-api/config"
	"notification-api/internal/controller"
	"notification-api/internal/repository"
	"notification-api/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	notificationMiddleware "notification-api/internal/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()

	notificationMiddleware.InitJWT(cfg.JWTSecret)

	db := config.NewDatabase(cfg)
	defer db.Close()

	redisClient := config.NewRedisClient(cfg)
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

	e.Use(notificationMiddleware.RequestIDMiddleware)
	e.Use(notificationMiddleware.LoggerMiddleware)
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
	auth.PUT("/notifications/:id/read", notificationController.MarkAsRead)
	auth.DELETE("/notifications/:id", notificationController.Delete)

	auth.POST("/webhooks", webhookController.Create)
	auth.GET("/webhooks", webhookController.List)
	auth.DELETE("/webhooks/:id", webhookController.Delete)

	admin := auth.Group("/admin")
	admin.Use(notificationMiddleware.AdminMiddleware)

	admin.GET("/users", adminController.ListUsers)
	admin.GET("/notifications", adminController.ListNotifications)

	go func() {
		if err := e.Start(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
