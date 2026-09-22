package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/ares/dp-vc-webApp/controllers/health"
)

// Health registers health-related routes
func Health(router *gin.Engine) {
	healthCtrl := health.New()

	router.GET("/health", healthCtrl.Health)
	router.GET("/health/ready", healthCtrl.Ready)
	router.GET("/health/live", healthCtrl.Live)
}
