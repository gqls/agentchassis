package billing

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handlers exposes billing over HTTP. Admin routes are mounted behind the
// auth-service admin middleware; the webhook route is public — its
// authentication is the provider signature, verified in the service.
type Handlers struct {
	service *Service
	logger  *zap.Logger
}

func NewHandlers(service *Service, logger *zap.Logger) *Handlers {
	return &Handlers{service: service, logger: logger}
}

type createVoucherRequest struct {
	DropsPriceToPence int    `json:"drops_price_to_pence" binding:"required"`
	RecipientName     string `json:"recipient_name"`
	ExpiresAt         string `json:"expires_at"`   // RFC3339; or use ttl_days
	TTLDays           int    `json:"ttl_days"`     // convenience alternative
}

// HandleCreateVoucher issues a single-use, named, expiring code (£10/£55).
func (h *Handlers) HandleCreateVoucher(c *gin.Context) {
	var req createVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var expires time.Time
	switch {
	case req.ExpiresAt != "":
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be RFC3339"})
			return
		}
		expires = t
	case req.TTLDays > 0:
		expires = time.Now().AddDate(0, 0, req.TTLDays)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide expires_at (RFC3339) or ttl_days"})
		return
	}
	v, err := h.service.CreateVoucher(c.Request.Context(), req.DropsPriceToPence, req.RecipientName, expires)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *Handlers) HandleListVouchers(c *gin.Context) {
	vouchers, err := h.service.ListVouchers(c.Request.Context())
	if err != nil {
		h.logger.Error("list vouchers failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list vouchers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vouchers": vouchers, "count": len(vouchers)})
}

type createOrderRequest struct {
	ClientID    string `json:"client_id" binding:"required"`
	VoucherCode string `json:"voucher_code"`
	Email       string `json:"email"` // overrides the client's stored email
}

// HandleCreateOrder creates an order + checkout link for a customer. Under
// the ruled after_approval timing the owner calls this from the dashboard at
// preview approval; under upfront the intake gate will call the same thing.
func (h *Handlers) HandleCreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, url, err := h.service.CreateOrder(c.Request.Context(), req.ClientID, req.VoucherCode, req.Email)
	switch {
	case errors.Is(err, ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": ErrNotConfigured.Error()})
		return
	case errors.Is(err, ErrVoucherInvalid):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": ErrVoucherInvalid.Error()})
		return
	case err != nil:
		h.logger.Error("create order failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"order": order, "checkout_url": url})
}

func (h *Handlers) HandleListOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	orders, err := h.service.ListOrders(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("list orders failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "count": len(orders)})
}

func (h *Handlers) HandleGetSettings(c *gin.Context) {
	timing, err := h.service.GetPaymentTiming(c.Request.Context())
	if err != nil {
		h.logger.Error("get billing settings failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read billing settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment_timing": timing, "configured": h.service.Configured()})
}

type updateSettingsRequest struct {
	PaymentTiming string `json:"payment_timing" binding:"required"`
}

func (h *Handlers) HandleUpdateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.SetPaymentTiming(c.Request.Context(), req.PaymentTiming); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment_timing": req.PaymentTiming})
}

// HandleStripeWebhook is the public endpoint. It reads the raw body (the
// signature covers the exact bytes) and hands verification to the service.
// It answers 200 for anything verified — including duplicates — because a
// non-2xx makes Stripe retry, and a verified duplicate needs no retry.
func (h *Handlers) HandleStripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable body"})
		return
	}
	err = h.service.HandleWebhook(c.Request.Context(), payload, c.GetHeader("Stripe-Signature"))
	switch {
	case errors.Is(err, ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": ErrNotConfigured.Error()})
	case err != nil:
		// Signature failures land here too: reject, log, no detail out.
		h.logger.Warn("webhook rejected", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "rejected"})
	default:
		c.JSON(http.StatusOK, gin.H{"received": true})
	}
}
