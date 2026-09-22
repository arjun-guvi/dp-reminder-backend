package env

import (
	"os"
	"testing"
)

func TestGetEnv_Default(t *testing.T) {
	value := GetEnv("NONEXISTENT_KEY", "default_value")
	if value != "default_value" {
		t.Errorf("Expected default_value, got %s", value)
	}
}

func TestGetEnv_Existing(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	value := GetEnv("TEST_KEY", "default_value")
	if value != "test_value" {
		t.Errorf("Expected test_value, got %s", value)
	}
}

func TestGetEnvInt_Default(t *testing.T) {
	value := GetEnvInt("NONEXISTENT_INT_KEY", 42)
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}
}

func TestGetEnvInt_Existing(t *testing.T) {
	os.Setenv("TEST_INT_KEY", "100")
	defer os.Unsetenv("TEST_INT_KEY")

	value := GetEnvInt("TEST_INT_KEY", 42)
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("TEST_INVALID_INT", "not_a_number")
	defer os.Unsetenv("TEST_INVALID_INT")

	value := GetEnvInt("TEST_INVALID_INT", 42)
	if value != 42 {
		t.Errorf("Expected default 42 for invalid input, got %d", value)
	}
}

func TestGetEnvBool_Default(t *testing.T) {
	value := GetEnvBool("NONEXISTENT_BOOL_KEY", true)
	if value != true {
		t.Errorf("Expected true, got %v", value)
	}
}

func TestGetEnvBool_Existing(t *testing.T) {
	os.Setenv("TEST_BOOL_KEY", "false")
	defer os.Unsetenv("TEST_BOOL_KEY")

	value := GetEnvBool("TEST_BOOL_KEY", true)
	if value != false {
		t.Errorf("Expected false, got %v", value)
	}
}

func TestLoad_CreatesConfig(t *testing.T) {
	// Set some test values
	os.Setenv("PORT", "9090")
	os.Setenv("APP_ENV", "testing")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("APP_ENV")
	}()

	// Load should work (even if .env doesn't exist)
	err := Load()
	if err != nil {
		t.Errorf("Load should not return error: %v", err)
	}

	config := GetConfig()
	if config == nil {
		t.Error("Config should not be nil after Load")
		return
	}

	if config.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", config.Port)
	}

	if config.AppEnv != "testing" {
		t.Errorf("Expected env testing, got %s", config.AppEnv)
	}
}
