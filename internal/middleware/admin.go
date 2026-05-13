package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, ok := c.Get("role").(string)
		if !ok || role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "admin access required",
			})
		}

		return next(c)
	}
}
