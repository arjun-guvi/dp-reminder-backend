package routes

import (
	configauth "github.com/ares/dp-vc-webApp/configs/auth"
	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/types"
	paymentscontroller "github.com/ares/dp-vc-webApp/controllers/payments"
	paymentservice "github.com/ares/dp-vc-webApp/services/payments"
	"github.com/gin-gonic/gin"
)

func Payments(router *gin.Engine) {
	controller := paymentscontroller.New(paymentservice.New(env.GetConfig()))
	group := router.Group("/payments")
	group.Use(configauth.Auth(types.RouterArr{}))
	group.POST("", controller.Create)
	group.GET("", controller.List)
	group.GET("/:id", controller.Get)
	group.PATCH("/:id/status", controller.UpdateStatus)
	group.DELETE("/:id", controller.Delete)
}
