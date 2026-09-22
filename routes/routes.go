package routes

import (
	"github.com/gin-gonic/gin"
)

// Register registers all application routes
func Register(router *gin.Engine) {
	// Inject route groups
	Health(router)
	Auth(router)
	Payments(router)
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"status": "error",
			"error":  "Route not found",
		})
	})
}
