package routes

import (
	"github.com/ares/dp-vc-webApp/configs/env"
	authcontroller "github.com/ares/dp-vc-webApp/controllers/auth"
	authservice "github.com/ares/dp-vc-webApp/services/auth"
	"github.com/gin-gonic/gin"
)

func Auth(router *gin.Engine) {
	controller := authcontroller.New(authservice.New(env.GetConfig()))
	group := router.Group("/auth")
	group.POST("/signup", controller.Signup)
	group.POST("/login", controller.Login)
	group.POST("/logout", controller.Logout)
}
