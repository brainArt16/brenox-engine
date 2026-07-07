package billing

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/brainart16/brenox/internal/httperr"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *Handler) GetPlatformStatus(c *gin.Context) {
	status, err := h.service.GetPlatformStatus(c.Request.Context())
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetAppBilling(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}
	userID := c.MustGet("user_id").(int64)
	if err := h.service.AssertAppOwner(c.Request.Context(), appID, userID); err != nil {
		writeError(c, err)
		return
	}

	billing, err := h.service.GetAppBilling(c.Request.Context(), appID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, billing)
}

func (h *Handler) CreateCheckout(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}
	userID := c.MustGet("user_id").(int64)
	if err := h.service.AssertAppOwner(c.Request.Context(), appID, userID); err != nil {
		writeError(c, err)
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid request body")
		return
	}

	email, err := h.service.UserEmail(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}

	checkout, err := h.service.CreateCheckoutSession(c.Request.Context(), userID, email, appID, req.PlanSlug)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, checkout)
}

func (h *Handler) StripeWebhook(c *gin.Context) {
	payload, err := ReadWebhookBody(c.Request, 65536)
	if err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid payload")
		return
	}
	sig := c.GetHeader("Stripe-Signature")
	if err := h.service.HandleStripeWebhook(c.Request.Context(), payload, sig); err != nil {
		if errors.Is(err, ErrStripeNotConfigured) {
			httperr.WriteJSON(c, http.StatusServiceUnavailable, httperr.ClientMessage(err, ErrStripeNotConfigured))
			return
		}
		httperr.WriteJSON(c, http.StatusBadRequest, "webhook verification failed")
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) AdminGetBillingOverview(c *gin.Context) {
	overview, err := h.service.AdminGetOverview(c.Request.Context())
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *Handler) AdminListSubscriptions(c *gin.Context) {
	limit, offset := pagination(c)
	items, err := h.service.AdminListSubscriptions(c.Request.Context(), limit, offset)
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": items})
}

func (h *Handler) AdminGetAppBilling(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}
	billing, err := h.service.GetAppBilling(c.Request.Context(), appID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, billing)
}

func (h *Handler) AdminUpdateAppSubscription(c *gin.Context) {
	appID, ok := parseID(c, "app_id")
	if !ok {
		return
	}
	var req AdminUpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid request body")
		return
	}
	billing, err := h.service.AdminUpdateSubscription(c.Request.Context(), appID, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, billing)
}

func (h *Handler) AdminGetPlatformSettings(c *gin.Context) {
	settings, err := h.service.GetPlatformSettings(c.Request.Context())
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) AdminUpdatePlatformSettings(c *gin.Context) {
	var req UpdatePlatformSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid request body")
		return
	}
	settings, err := h.service.UpdatePlatformSettings(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *Handler) AdminListPlans(c *gin.Context) {
	plans, err := h.service.AdminListPlans(c.Request.Context())
	if err != nil {
		httperr.WriteInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *Handler) AdminCreatePlan(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid request body")
		return
	}
	plan, err := h.service.AdminCreatePlan(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *Handler) AdminUpdatePlan(c *gin.Context) {
	slug := c.Param("slug")
	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid request body")
		return
	}
	plan, err := h.service.AdminUpdatePlan(c.Request.Context(), slug, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *Handler) AdminDeletePlan(c *gin.Context) {
	slug := c.Param("slug")
	if err := h.service.AdminDeletePlan(c.Request.Context(), slug); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func pagination(c *gin.Context) (int32, int32) {
	limit := int32(50)
	offset := int32(0)
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}
	if raw := c.Query("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}
	return limit, offset
}

func parseID(c *gin.Context, param string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		httperr.WriteJSON(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httperr.WriteJSON(c, http.StatusNotFound, httperr.ClientMessage(err, ErrNotFound))
	case errors.Is(err, ErrForbidden):
		httperr.WriteJSON(c, http.StatusForbidden, httperr.ClientMessage(err, ErrForbidden))
	case errors.Is(err, ErrInvalidPlan), errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrInvalidSlug):
		httperr.WriteJSON(c, http.StatusBadRequest, httperr.ClientMessage(err, ErrInvalidPlan, ErrInvalidRequest, ErrInvalidSlug))
	case errors.Is(err, ErrSlugTaken):
		httperr.WriteJSON(c, http.StatusConflict, httperr.ClientMessage(err, ErrSlugTaken))
	case errors.Is(err, ErrPlanInUse):
		httperr.WriteJSON(c, http.StatusConflict, httperr.ClientMessage(err, ErrPlanInUse))
	case errors.Is(err, ErrStripeNotConfigured):
		httperr.WriteJSON(c, http.StatusServiceUnavailable, httperr.ClientMessage(err, ErrStripeNotConfigured))
	case errors.Is(err, ErrPlanStripePriceMissing):
		httperr.WriteJSON(c, http.StatusServiceUnavailable, httperr.ClientMessage(err, ErrPlanStripePriceMissing))
	case errors.Is(err, ErrMessageLimit):
		httperr.WriteJSON(c, http.StatusPaymentRequired, httperr.ClientMessage(err, ErrMessageLimit))
	case errors.Is(err, ErrSubscriptionInactive):
		httperr.WriteJSON(c, http.StatusPaymentRequired, httperr.ClientMessage(err, ErrSubscriptionInactive))
	case errors.Is(err, ErrWebhooksNotAllowed), errors.Is(err, ErrVideoNotAllowed):
		httperr.WriteJSON(c, http.StatusPaymentRequired, httperr.ClientMessage(err, ErrWebhooksNotAllowed, ErrVideoNotAllowed))
	default:
		httperr.WriteInternal(c, err)
	}
}
