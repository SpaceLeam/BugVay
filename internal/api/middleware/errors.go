package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, APIError{
					Code:    http.StatusBadRequest,
					Message: "Invalid request format",
					Details: err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusBadRequest, APIError{
					Code:    http.StatusBadRequest,
					Message: err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, APIError{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error",
					Details: err.Error(),
				})
			}
		}
	}
}

func AbortWithError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, APIError{
		Code:    code,
		Message: message,
	})
}
