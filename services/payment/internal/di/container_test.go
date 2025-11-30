package di

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	container := NewContainer()
	assert.NotNil(t, container)
}

func TestContainer_Config(t *testing.T) {
	container := NewContainer()

	// Первый вызов должен инициализировать конфигурацию
	cfg1 := container.Config()
	assert.NotNil(t, cfg1)

	// Второй вызов должен вернуть тот же экземпляр
	cfg2 := container.Config()
	assert.Equal(t, cfg1, cfg2)
}

func TestContainer_Service(t *testing.T) {
	container := NewContainer()
	ctx := context.Background()

	// Первый вызов должен инициализировать сервис
	service1, err := container.Service(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, service1)

	// Второй вызов должен вернуть тот же экземпляр
	service2, err := container.Service(ctx)
	assert.NoError(t, err)
	assert.Equal(t, service1, service2)
}

func TestContainer_Close(t *testing.T) {
	container := NewContainer()
	err := container.Close()
	assert.NoError(t, err)
}

func TestContainer_ConfigMultipleCalls(t *testing.T) {
	container := NewContainer()

	// Множественные вызовы должны возвращать тот же экземпляр
	cfg1 := container.Config()
	cfg2 := container.Config()
	cfg3 := container.Config()

	assert.Equal(t, cfg1, cfg2)
	assert.Equal(t, cfg2, cfg3)
}

func TestContainer_ServiceMultipleCalls(t *testing.T) {
	container := NewContainer()
	ctx := context.Background()

	// Множественные вызовы должны возвращать тот же экземпляр
	service1, err1 := container.Service(ctx)
	service2, err2 := container.Service(ctx)
	service3, err3 := container.Service(ctx)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, service1, service2)
	assert.Equal(t, service2, service3)
}
