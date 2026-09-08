package handler

import (
	"errors"
	"net/http"

	"core-service/internal/domain"

	"github.com/gin-gonic/gin"
)

// ErrorResponse представляет унифицированный ответ с ошибкой.
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

// SuccessResponse представляет стандартный ответ об успешном выполнении действия.
type SuccessResponse struct {
	Message string `json:"message"`
}

// RespondError обрабатывает доменные и системные ошибки и отправляет соответствующий HTTP-статус.
func RespondError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrItemInStopList),
		errors.Is(err, domain.ErrOutOfStock),
		errors.Is(err, domain.ErrCannotCancelOrder),
		errors.Is(err, domain.ErrInvalidStatusTransition):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidInput),
		errors.Is(err, domain.ErrEmptyOrder),
		errors.Is(err, domain.ErrRestaurantInactive),
		errors.Is(err, domain.ErrPartnerIDMismatch),
		errors.Is(err, domain.ErrRestaurantIDRequired):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal server error",
			Details: []string{err.Error()},
		})
	}
}

// RespondSuccess отправляет стандартный ответ об успешном выполнении действия.
func RespondSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusOK, SuccessResponse{Message: message})
}
