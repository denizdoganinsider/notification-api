package controller

import (
	"net/http"
	"notification-api/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

type AdminController struct {
	userService         *service.UserService
	notificationService *service.NotificationService
}

func NewAdminController(
	userService *service.UserService,
	notificationService *service.NotificationService,
) *AdminController {
	return &AdminController{
		userService:         userService,
		notificationService: notificationService,
	}
}

func (ac *AdminController) ListUsers(c echo.Context) error {
	users, err := ac.userService.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch users",
		})
	}

	return c.JSON(http.StatusOK, users)
}

func (ac *AdminController) ListNotifications(c echo.Context) error {
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

	result, err := ac.notificationService.ListAll(page, perPage)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch notifications",
		})
	}

	return c.JSON(http.StatusOK, result)
}
