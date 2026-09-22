package auth

import (
	"strings"

	"github.com/ares/dp-vc-webApp/configs/response"
	"github.com/ares/dp-vc-webApp/models/user"
	authservice "github.com/ares/dp-vc-webApp/services/auth"
	"github.com/gin-gonic/gin"
)

type Controller struct{ service *authservice.Service }

func New(service *authservice.Service) *Controller { return &Controller{service: service} }

func (ctrl *Controller) Signup(c *gin.Context) {
	var request user.SignupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	result, err := ctrl.service.Signup(c.Request.Context(), request)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, result)
}

func (ctrl *Controller) Login(c *gin.Context) {
	var request user.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	result, err := ctrl.service.Login(c.Request.Context(), request)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (ctrl *Controller) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "missing authorization header")
		return
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Unauthorized(c, "invalid authorization header format")
		return
	}
	token := parts[1]

	// Logout the user (blacklist the token)
	err := ctrl.service.Logout(c.Request.Context(), token)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "successfully logged out"})
}
