package httpapi

import (
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"example.com/payment-gateway/internal/application/payment"
	"github.com/gin-gonic/gin"
)

func NewRouter(paymentService payment.Service) http.Handler {
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	payments := paymentHandler{service: paymentService}
	v1 := router.Group("/v1")
	v1.POST("/payments", payments.Create)
	v1.GET("/payments/:id", payments.Get)

	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))

	return router
}
