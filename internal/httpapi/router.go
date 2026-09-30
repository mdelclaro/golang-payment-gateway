package httpapi

import (
	"net/http"

	"example.com/payment-gateway/internal/application/payment"
	"github.com/gin-gonic/gin"
)

func NewRouter(paymentService payment.Service) http.Handler {
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/payments", createPaymentHandler{service: paymentService}.Create)

	return router
}
