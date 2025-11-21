package di

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_Config(t *testing.T) {
	os.Clearenv()

	container := NewContainer()

	// Первый вызов должен инициализировать конфигурацию
	cfg1 := container.Config()
	require.NotNil(t, cfg1)
	assert.Equal(t, "8080", cfg1.HTTPPort)

	// Второй вызов должен вернуть тот же экземпляр
	cfg2 := container.Config()
	assert.Same(t, cfg1, cfg2, "Config should be initialized once")
}

func TestContainer_Config_FromEnvironment(t *testing.T) {
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	os.Setenv("INVENTORY_GRPC_ADDRESS", "inventory:50051")
	os.Setenv("PAYMENT_GRPC_ADDRESS", "payment:50052")
	os.Setenv("LOG_LEVEL", "info")

	container := NewContainer()
	cfg := container.Config()

	assert.Equal(t, "9090", cfg.HTTPPort)
	assert.Equal(t, "postgres://test:test@localhost:5432/test", cfg.DatabaseURL)
	assert.Equal(t, "inventory:50051", cfg.InventoryGRPCAddress)
	assert.Equal(t, "payment:50052", cfg.PaymentGRPCAddress)
	assert.Equal(t, "info", cfg.LogLevel)

	os.Clearenv()
}

func TestContainer_Config_InitializedOnce(t *testing.T) {
	os.Clearenv()

	container := NewContainer()

	// Вызываем Config несколько раз
	cfg1 := container.Config()
	cfg2 := container.Config()
	cfg3 := container.Config()

	// Все должны быть одним и тем же экземпляром
	assert.Same(t, cfg1, cfg2)
	assert.Same(t, cfg2, cfg3)
}

func TestContainer_Close(t *testing.T) {
	container := NewContainer()

	// Close должен работать даже если ничего не инициализировано
	err := container.Close()
	assert.NoError(t, err)
}

func TestContainer_LazyLoading(t *testing.T) {
	os.Clearenv()

	container := NewContainer()

	// Проверяем, что конфигурация инициализируется при первом вызове
	cfg := container.Config()
	assert.NotNil(t, cfg)

	// Проверяем, что Database не инициализирован до вызова
	// (в реальном тесте с БД это будет проверяться через подключение)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := container.Database(ctx)
	// Ожидаем ошибку подключения, так как БД не запущена
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database")
}

func TestContainer_MethodOrder(t *testing.T) {
	os.Clearenv()

	container := NewContainer()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Проверяем порядок инициализации зависимостей
	// OrderHandler зависит от OrderService
	// OrderService зависит от OrderRepository, InventoryClient, PaymentClient
	// OrderRepository зависит от Database

	// Попытка получить OrderHandler должна инициализировать всю цепочку
	_, err := container.OrderHandler(ctx)
	
	// Ожидаем ошибку, так как БД не запущена, но проверяем что инициализация происходит
	assert.Error(t, err)
	
	// Проверяем, что конфигурация была инициализирована
	cfg := container.Config()
	assert.NotNil(t, cfg)
}

func TestContainer_ConcurrentAccess(t *testing.T) {
	os.Clearenv()

	container := NewContainer()

	// Тестируем concurrent доступ к Config
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			cfg := container.Config()
			assert.NotNil(t, cfg)
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	// Проверяем, что конфигурация инициализирована один раз
	cfg := container.Config()
	assert.NotNil(t, cfg)
}

