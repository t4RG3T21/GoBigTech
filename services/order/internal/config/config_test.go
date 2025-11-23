package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Очищаем переменные окружения для чистого теста
	os.Clearenv()

	cfg := Load()

	if cfg.HTTPPort != "8080" {
		t.Errorf("Expected HTTPPort to be '8080', got '%s'", cfg.HTTPPort)
	}

	expectedDBURL := "postgres://postgres:password@localhost:5432/order_service?sslmode=disable"
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("Expected DatabaseURL to be '%s', got '%s'", expectedDBURL, cfg.DatabaseURL)
	}

	if cfg.InventoryGRPCAddress != "localhost:50051" {
		t.Errorf("Expected InventoryGRPCAddress to be 'localhost:50051', got '%s'", cfg.InventoryGRPCAddress)
	}

	if cfg.PaymentGRPCAddress != "localhost:50052" {
		t.Errorf("Expected PaymentGRPCAddress to be 'localhost:50052', got '%s'", cfg.PaymentGRPCAddress)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestLoad_FromEnvironment(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5433/test_db")
	os.Setenv("INVENTORY_GRPC_ADDRESS", "inventory:50051")
	os.Setenv("PAYMENT_GRPC_ADDRESS", "payment:50052")
	os.Setenv("LOG_LEVEL", "info")

	cfg := Load()

	if cfg.HTTPPort != "9090" {
		t.Errorf("Expected HTTPPort to be '9090', got '%s'", cfg.HTTPPort)
	}

	if cfg.DatabaseURL != "postgres://user:pass@localhost:5433/test_db" {
		t.Errorf("Expected DatabaseURL to be 'postgres://user:pass@localhost:5433/test_db', got '%s'", cfg.DatabaseURL)
	}

	if cfg.InventoryGRPCAddress != "inventory:50051" {
		t.Errorf("Expected InventoryGRPCAddress to be 'inventory:50051', got '%s'", cfg.InventoryGRPCAddress)
	}

	if cfg.PaymentGRPCAddress != "payment:50052" {
		t.Errorf("Expected PaymentGRPCAddress to be 'payment:50052', got '%s'", cfg.PaymentGRPCAddress)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	// Очищаем после теста
	os.Clearenv()
}

func TestGetEnv(t *testing.T) {
	os.Clearenv()

	// Тест значения по умолчанию
	value := getEnv("NON_EXISTENT_VAR", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}

	// Тест с установленной переменной
	os.Setenv("TEST_VAR", "test_value")
	value = getEnv("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	os.Clearenv()
}

func TestGetEnvAsInt(t *testing.T) {
	os.Clearenv()

	// Тест значения по умолчанию
	value := getEnvAsInt("NON_EXISTENT_VAR", 42)
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Тест с валидным int
	os.Setenv("TEST_INT", "100")
	value = getEnvAsInt("TEST_INT", 42)
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}

	// Тест с невалидным int (должен вернуть default)
	os.Setenv("TEST_INT", "not_a_number")
	value = getEnvAsInt("TEST_INT", 42)
	if value != 42 {
		t.Errorf("Expected 42 (default for invalid), got %d", value)
	}

	os.Clearenv()
}

func TestGetEnvAsDuration(t *testing.T) {
	os.Clearenv()

	// Тест значения по умолчанию
	defaultDuration := 5 * time.Second
	value := getEnvAsDuration("NON_EXISTENT_VAR", defaultDuration)
	if value != defaultDuration {
		t.Errorf("Expected %v, got %v", defaultDuration, value)
	}

	// Тест с валидным duration
	os.Setenv("TEST_DURATION", "10s")
	value = getEnvAsDuration("TEST_DURATION", defaultDuration)
	if value != 10*time.Second {
		t.Errorf("Expected 10s, got %v", value)
	}

	// Тест с невалидным duration (должен вернуть default)
	os.Setenv("TEST_DURATION", "invalid")
	value = getEnvAsDuration("TEST_DURATION", defaultDuration)
	if value != defaultDuration {
		t.Errorf("Expected %v (default for invalid), got %v", defaultDuration, value)
	}

	os.Clearenv()
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				HTTPPort:              "8080",
				DatabaseURL:           "postgres://localhost/db",
				InventoryGRPCAddress:  "localhost:50051",
				PaymentGRPCAddress:    "localhost:50052",
				KafkaBootstrapServers: "localhost:9092",
				PaymentTopic:          "orders.payment",
				AssemblyTopic:         "orders.assembly",
				ConsumerGroupID:       "order-service",
				LogLevel:              "debug",
			},
			wantErr: false,
		},
		{
			name: "empty HTTPPort",
			config: &Config{
				HTTPPort:              "",
				DatabaseURL:           "postgres://localhost/db",
				InventoryGRPCAddress:  "localhost:50051",
				PaymentGRPCAddress:    "localhost:50052",
				KafkaBootstrapServers: "localhost:9092",
				PaymentTopic:          "orders.payment",
				AssemblyTopic:         "orders.assembly",
				ConsumerGroupID:       "order-service",
			},
			wantErr: true,
		},
		{
			name: "empty DatabaseURL",
			config: &Config{
				HTTPPort:              "8080",
				DatabaseURL:           "",
				InventoryGRPCAddress:  "localhost:50051",
				PaymentGRPCAddress:    "localhost:50052",
				KafkaBootstrapServers: "localhost:9092",
				PaymentTopic:          "orders.payment",
				AssemblyTopic:         "orders.assembly",
				ConsumerGroupID:       "order-service",
			},
			wantErr: true,
		},
		{
			name: "empty InventoryGRPCAddress",
			config: &Config{
				HTTPPort:              "8080",
				DatabaseURL:           "postgres://localhost/db",
				InventoryGRPCAddress:  "",
				PaymentGRPCAddress:    "localhost:50052",
				KafkaBootstrapServers: "localhost:9092",
				PaymentTopic:          "orders.payment",
				AssemblyTopic:         "orders.assembly",
				ConsumerGroupID:       "order-service",
			},
			wantErr: true,
		},
		{
			name: "empty PaymentGRPCAddress",
			config: &Config{
				HTTPPort:              "8080",
				DatabaseURL:           "postgres://localhost/db",
				InventoryGRPCAddress:  "localhost:50051",
				PaymentGRPCAddress:    "",
				KafkaBootstrapServers: "localhost:9092",
				PaymentTopic:          "orders.payment",
				AssemblyTopic:         "orders.assembly",
				ConsumerGroupID:       "order-service",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
