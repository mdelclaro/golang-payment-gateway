package httpapi

import (
	"net/http"

	"example.com/payment-gateway/internal/application/payment"
	"github.com/gin-gonic/gin"
)

func NewRouter() http.Handler {
	return NewRouterWithPayments(nil)
}

func NewRouterWithPayments(paymentService payment.Service) http.Handler {
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	if paymentService != nil {
		router.POST("/payments", createPaymentHandler{service: paymentService}.Create)
	}

	return router
}
