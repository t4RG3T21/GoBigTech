package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Сохраняем оригинальные значения
	originalPort := os.Getenv("GRPC_PORT")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer func() {
		if originalPort != "" {
			os.Setenv("GRPC_PORT", originalPort)
		} else {
			os.Unsetenv("GRPC_PORT")
		}
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	t.Run("default values", func(t *testing.T) {
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("LOG_LEVEL")

		cfg := Load()
		assert.Equal(t, "50052", cfg.GRPCPort)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("from environment", func(t *testing.T) {
		os.Setenv("GRPC_PORT", "9999")
		os.Setenv("LOG_LEVEL", "info")

		cfg := Load()
		assert.Equal(t, "9999", cfg.GRPCPort)
		assert.Equal(t, "info", cfg.LogLevel)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			GRPCPort: "50052",
			LogLevel: "debug",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty port", func(t *testing.T) {
		cfg := &Config{
			GRPCPort: "",
			LogLevel: "debug",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GRPC_PORT cannot be empty")
	})

	t.Run("empty log level is ok", func(t *testing.T) {
		cfg := &Config{
			GRPCPort: "50052",
			LogLevel: "",
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

func TestGetEnv(t *testing.T) {
	originalValue := os.Getenv("TEST_ENV_VAR")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_ENV_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_ENV_VAR")
		}
	}()

	t.Run("returns environment value", func(t *testing.T) {
		os.Setenv("TEST_ENV_VAR", "test-value")
		result := getEnv("TEST_ENV_VAR", "default")
		assert.Equal(t, "test-value", result)
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_ENV_VAR")
		result := getEnv("TEST_ENV_VAR", "default-value")
		assert.Equal(t, "default-value", result)
	})
}

func TestGetEnvAsInt(t *testing.T) {
	originalValue := os.Getenv("TEST_INT_VAR")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_INT_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_INT_VAR")
		}
	}()

	t.Run("returns parsed int", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR", "42")
		result := getEnvAsInt("TEST_INT_VAR", 0)
		assert.Equal(t, 42, result)
	})

	t.Run("returns default on invalid value", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR", "not-a-number")
		result := getEnvAsInt("TEST_INT_VAR", 100)
		assert.Equal(t, 100, result)
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_INT_VAR")
		result := getEnvAsInt("TEST_INT_VAR", 200)
		assert.Equal(t, 200, result)
	})
}

func TestGetEnvAsDuration(t *testing.T) {
	originalValue := os.Getenv("TEST_DURATION_VAR")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_DURATION_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_DURATION_VAR")
		}
	}()

	t.Run("returns parsed duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION_VAR", "1h30m")
		result := getEnvAsDuration("TEST_DURATION_VAR", 0)
		assert.Equal(t, "1h30m0s", result.String())
	})

	t.Run("returns default on invalid value", func(t *testing.T) {
		os.Setenv("TEST_DURATION_VAR", "invalid")
		result := getEnvAsDuration("TEST_DURATION_VAR", 0)
		assert.Equal(t, 0, int(result))
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_DURATION_VAR")
		result := getEnvAsDuration("TEST_DURATION_VAR", 0)
		assert.Equal(t, 0, int(result))
	})
}
