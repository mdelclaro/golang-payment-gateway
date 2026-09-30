package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	applicationpayment "example.com/payment-gateway/internal/application/payment"
	"example.com/payment-gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

type paymentHandler struct {
	service applicationpayment.Service
}

func (h paymentHandler) Create(c *gin.Context) {
	var request createPaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid payment request"})
		return
	}

	created, err := h.service.CreatePayment(c.Request.Context(), applicationpayment.CreatePaymentInput{
		CardNumber:  request.CardNumber,
		ExpiryMonth: request.ExpiryMonth,
		ExpiryYear:  request.ExpiryYear,
		Currency:    request.Currency,
		Amount:      request.Amount,
		CVV:         request.CVV,
	})
	if err != nil {
		if errors.Is(err, applicationpayment.ErrInvalidPayment) {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid payment request"})
			return
		}

		if errors.Is(err, applicationpayment.ErrBankUnavailable) {
			c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "payment authorization is unavailable"})
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse{Error: "payment could not be processed"})
		return
	}

	c.JSON(http.StatusCreated, newPaymentResponse(created))
}

func (h paymentHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid payment ID"})
		return
	}

	found, err := h.service.GetPayment(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, applicationpayment.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "payment not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse{Error: "payment could not be retrieved"})
		return
	}

	c.JSON(http.StatusOK, newPaymentResponse(found))
}

func newPaymentResponse(payment domain.Payment) paymentResponse {
	return paymentResponse{
		ID:                payment.ID,
		Status:            string(payment.Status),
		CardLastFour:      payment.CardLastFour,
		ExpiryMonth:       payment.ExpiryMonth,
		ExpiryYear:        payment.ExpiryYear,
		Currency:          payment.Currency,
		Amount:            payment.Amount,
		AuthorizationCode: payment.AuthorizationCode,
	}
}
