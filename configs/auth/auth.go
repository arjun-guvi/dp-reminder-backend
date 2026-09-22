package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/response"
	"github.com/ares/dp-vc-webApp/configs/types"
	authservice "github.com/ares/dp-vc-webApp/services/auth"
)

const (
	// ContextKeyUser is the key for storing user in context
	ContextKeyUser = "user"
)

// Auth returns an authentication middleware
// routerPermissions maps route paths to required permissions
func Auth(routerPermissions types.RouterArr) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, "Token required")
			c.Abort()
			return
		}

		// Validate token and get user
		user, err := ValidateToken(c, token)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Check permissions for current route
		if err := checkPermissions(c, routerPermissions, user); err != nil {
			response.Forbidden(c, err.Error())
			c.Abort()
			return
		}

		// Store user in context
		c.Set(ContextKeyUser, user)
		c.Next()
	}
}

// ValidateToken validates a token and returns the user
// This is a placeholder implementation - replace with actual JWT validation
func ValidateToken(c *gin.Context, token string) (*types.User, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}
	return authservice.New(env.GetConfig()).ValidateToken(token)
}

// checkPermissions verifies user has required permissions for the route
func checkPermissions(c *gin.Context, routerPermissions types.RouterArr, user *types.User) error {
	// Get current route
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}

	// Check if route requires specific permissions
	if requiredPerms, exists := routerPermissions[route]; exists {
		if !hasPermissions(user.Permissions, requiredPerms) {
			return fmt.Errorf("insufficient permissions")
		}
	}

	return nil
}

// hasPermissions checks if user has all required permissions
func hasPermissions(userPerms, requiredPerms []string) bool {
	permMap := make(map[string]bool)
	for _, p := range userPerms {
		permMap[p] = true
	}

	for _, req := range requiredPerms {
		if !permMap[req] {
			return false
		}
	}

	return true
}

// GetUser retrieves the authenticated user from context
func GetUser(c *gin.Context) (*types.User, bool) {
	userVal, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil, false
	}

	user, ok := userVal.(*types.User)
	return user, ok
}

// MustGetUser retrieves user or panics if not found
func MustGetUser(c *gin.Context) *types.User {
	user, ok := GetUser(c)
	if !ok {
		panic("user not found in context - is auth middleware configured?")
	}
	return user
}

// OptionalAuthentication is a middleware that extracts user info if present
// but doesn't require authentication
func OptionalAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}

		token := parts[1]
		user, err := ValidateToken(c, token)
		if err == nil && user != nil {
			c.Set(ContextKeyUser, user)
		}

		c.Next()
	}
}

func init() {
}
