package controller

import (
	"database/sql"
	"net/http"
	"notification-api/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

type CreateNotificationRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type NotificationController struct {
	notificationService *service.NotificationService
}

func NewNotificationController(notificationService *service.NotificationService) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
	}
}

func (nc *NotificationController) Create(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	var req CreateNotificationRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	notification, err := nc.notificationService.Create(userID, req.Title, req.Message)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, notification)
}

func (nc *NotificationController) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	page := 1
	perPage := 10

	pageParam := c.QueryParam("page")
	if pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid page parameter",
			})
		}

		page = parsedPage
	}

	perPageParam := c.QueryParam("per_page")
	if perPageParam != "" {
		parsedPerPage, err := strconv.Atoi(perPageParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid per_page parameter",
			})
		}

		perPage = parsedPerPage
	}

	notifications, err := nc.notificationService.List(userID, page, perPage)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch notifications",
		})
	}

	return c.JSON(http.StatusOK, notifications)
}

func (nc *NotificationController) GetByID(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	idParam := c.Param("id")
	notificationID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid notification id",
		})
	}

	notification, err := nc.notificationService.GetByID(notificationID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "notification not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch notification",
		})
	}

	return c.JSON(http.StatusOK, notification)
}
