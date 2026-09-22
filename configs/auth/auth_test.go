package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ares/dp-vc-webApp/configs/types"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuth_MissingHeader(t *testing.T) {
	router := gin.New()
	router.GET("/protected", Auth(types.RouterArr{}), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	router := gin.New()
	router.GET("/protected", Auth(types.RouterArr{}), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	router := gin.New()
	router.GET("/protected", Auth(types.RouterArr{}), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetUser_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	user, exists := GetUser(c)
	if exists {
		t.Error("Expected user to not exist")
	}
	if user != nil {
		t.Error("Expected user to be nil")
	}
}

func TestGetUser_Set(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	expectedUser := &types.User{
		ID:    "123",
		Email: "test@example.com",
	}
	c.Set(ContextKeyUser, expectedUser)

	user, exists := GetUser(c)
	if !exists {
		t.Error("Expected user to exist")
	}
	if user == nil {
		t.Error("Expected user to not be nil")
		return
	}
	if user.ID != expectedUser.ID {
		t.Errorf("Expected user ID %s, got %s", expectedUser.ID, user.ID)
	}
}

func TestHasPermissions(t *testing.T) {
	userPerms := []string{"read", "write", "admin"}

	// Test with all permissions present
	if !hasPermissions(userPerms, []string{"read"}) {
		t.Error("Should have read permission")
	}

	if !hasPermissions(userPerms, []string{"read", "write"}) {
		t.Error("Should have read and write permissions")
	}

	// Test with missing permission
	if hasPermissions(userPerms, []string{"delete"}) {
		t.Error("Should not have delete permission")
	}
}
