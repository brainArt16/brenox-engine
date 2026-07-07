package billing

import (
	"errors"
	"net/http"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/gin-gonic/gin"
)

func WriteHTTPError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, ErrMessageLimit),
		errors.Is(err, ErrSubscriptionInactive),
		errors.Is(err, ErrWebhooksNotAllowed),
		errors.Is(err, ErrVideoNotAllowed):
		httperr.WriteJSON(c, http.StatusPaymentRequired, httperr.ClientMessage(
			err,
			ErrMessageLimit,
			ErrSubscriptionInactive,
			ErrWebhooksNotAllowed,
			ErrVideoNotAllowed,
		))
		return true
	default:
		return false
	}
}
