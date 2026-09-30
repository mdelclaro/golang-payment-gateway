package httpapi

import (
	"errors"
	"net/http"

	applicationpayment "example.com/payment-gateway/internal/application/payment"
	"github.com/gin-gonic/gin"
)

type createPaymentHandler struct {
	service applicationpayment.Service
}

func (h createPaymentHandler) Create(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "payment could not be processed"})
		return
	}

	c.JSON(http.StatusCreated, paymentResponse{
		ID:                created.ID,
		Status:            string(created.Status),
		CardLastFour:      created.CardLastFour,
		ExpiryMonth:       created.ExpiryMonth,
		ExpiryYear:        created.ExpiryYear,
		Currency:          created.Currency,
		Amount:            created.Amount,
		AuthorizationCode: created.AuthorizationCode,
	})
}
