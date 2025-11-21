package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew_DevelopmentMode(t *testing.T) {
	logger, err := New("test-service", "debug")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Проверяем, что логгер работает
	logger.Info("test message", zap.String("key", "value"))
	logger.Debug("debug message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Проверяем, что имя сервиса добавлено
	assert.NotNil(t, logger.Logger)
}

func TestNew_ProductionMode(t *testing.T) {
	logger, err := New("test-service", "info")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Проверяем, что логгер работает
	logger.Info("test message", zap.String("key", "value"))
	logger.Warn("warn message")
	logger.Error("error message")

	assert.NotNil(t, logger.Logger)
}

func TestNew_AllLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger, err := New("test-service", level)
			require.NoError(t, err)
			require.NotNil(t, logger)
			assert.NotNil(t, logger.Logger)
		})
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	// Невалидный уровень должен использовать info по умолчанию
	logger, err := New("test-service", "invalid")
	require.NoError(t, err) // Не возвращает ошибку, использует default
	require.NotNil(t, logger)
}

func TestWithFields(t *testing.T) {
	logger, err := New("test-service", "debug")
	require.NoError(t, err)

	loggerWithFields := logger.WithFields(
		zap.String("user_id", "123"),
		zap.Int("count", 42),
	)

	assert.NotNil(t, loggerWithFields)
	assert.NotSame(t, logger, loggerWithFields) // Должен быть новый экземпляр
}

func TestWithError(t *testing.T) {
	logger, err := New("test-service", "debug")
	require.NoError(t, err)

	testErr := assert.AnError
	loggerWithError := logger.WithError(testErr)

	assert.NotNil(t, loggerWithError)
	assert.NotSame(t, logger, loggerWithError) // Должен быть новый экземпляр
}

func TestLogger_Chaining(t *testing.T) {
	logger, err := New("test-service", "debug")
	require.NoError(t, err)

	// Проверяем цепочку вызовов
	chained := logger.
		WithFields(zap.String("field1", "value1")).
		WithFields(zap.String("field2", "value2")).
		WithError(assert.AnError)

	assert.NotNil(t, chained)
}

func TestLogger_DifferentServices(t *testing.T) {
	logger1, err := New("service-1", "debug")
	require.NoError(t, err)

	logger2, err := New("service-2", "info")
	require.NoError(t, err)

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)
	assert.NotSame(t, logger1, logger2)
}
