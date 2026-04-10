package controller

import (
	"net/http"
	"notification-api/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

type CreateWebhookRequest struct {
	URL string `json:"url"`
}

type WebhookController struct {
	webhookService *service.WebhookService
}

func NewWebhookController(webhookService *service.WebhookService) *WebhookController {
	return &WebhookController{
		webhookService: webhookService,
	}
}

func (wc *WebhookController) Create(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	var req CreateWebhookRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	webhook, err := wc.webhookService.Create(userID, req.URL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, webhook)
}

func (wc *WebhookController) List(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	webhooks, err := wc.webhookService.List(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch webhooks",
		})
	}

	return c.JSON(http.StatusOK, webhooks)
}

func (wc *WebhookController) Delete(c echo.Context) error {
	userID, ok := c.Get("user_id").(int64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	idParam := c.Param("id")
	webhookID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid webhook id",
		})
	}

	err = wc.webhookService.Delete(webhookID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to delete webhook",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "webhook deleted successfully",
	})
}
