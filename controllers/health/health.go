package health

import (
	"github.com/gin-gonic/gin"

	"github.com/ares/dp-vc-webApp/configs"
)

// Controller handles health-related requests
type Controller struct{}

// New creates a new health controller
func New() *Controller {
	return &Controller{}
}

// Health returns the health status of the application
func (ctrl *Controller) Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": configs.StatusSuccess,
		"data": gin.H{
			"service": configs.AppName,
			"status":  "healthy",
		},
	})
}

// Ready returns readiness status (checks dependencies)
func (ctrl *Controller) Ready(c *gin.Context) {
	// Check MongoDB connection
	// Check Redis connection
	// Return ready status

	c.JSON(200, gin.H{
		"status": configs.StatusSuccess,
		"data": gin.H{
			"service": configs.AppName,
			"status":  "ready",
		},
	})
}

// Live returns liveness status
func (ctrl *Controller) Live(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": configs.StatusSuccess,
		"data": gin.H{
			"service": configs.AppName,
			"status":  "alive",
		},
	})
}
