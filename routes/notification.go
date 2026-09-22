package routes

import (
	"github.com/ares/dp-vc-webApp/configs/auth"
	"github.com/ares/dp-vc-webApp/configs/types"
	notificationcontroller "github.com/ares/dp-vc-webApp/controllers/notification"
	"github.com/gin-gonic/gin"
)

// Notification registers notification-related routes
func Notification(router *gin.Engine) {
	controller := notificationcontroller.New()

	// Public routes (for testing)
	group := router.Group("/notifications")
	{
		// GET /notifications/status - Get worker status and configuration
		group.GET("/status", controller.GetWorkerStatus)

		// POST /notifications/test - Dry-run test notification check
		group.POST("/test", controller.TestNotifications)
	}

	// Protected routes (require authentication)
	protected := router.Group("/notifications")
	protected.Use(auth.Auth(types.RouterArr{
		"/notifications/trigger": {"notifications.manage"},
	}))
	{
		// POST /notifications/trigger - Actually trigger notifications
		protected.POST("/trigger", controller.TriggerNotifications)

		// DELETE /notifications/lock - Clear notification lock
		protected.DELETE("/lock", controller.ClearLock)
	}
}
