package di

import (
	"context"
	"sync"

	"github.com/t4RG3T21/GoBigTech/services/payment/internal/config"
	paymentservice "github.com/t4RG3T21/GoBigTech/services/payment/internal/service"
)

// Container - DI контейнер для Payment Service с ленивой загрузкой зависимостей
type Container struct {
	// Конфигурация инициализируется один раз через sync.Once
	configOnce sync.Once
	cfg        *config.Config

	// Сервис инициализируется один раз
	serviceOnce sync.Once
	service     *paymentservice.PaymentService
}

// NewContainer создает новый DI контейнер
func NewContainer() *Container {
	return &Container{}
}

// Config возвращает конфигурацию, инициализируется один раз
func (c *Container) Config() *config.Config {
	c.configOnce.Do(func() {
		c.cfg = config.Load()
	})
	return c.cfg
}

// Service возвращает сервис платежей с ленивой инициализацией
func (c *Container) Service(ctx context.Context) (*paymentservice.PaymentService, error) {
	c.serviceOnce.Do(func() {
		c.service = paymentservice.NewPaymentService()
	})

	return c.service, nil
}

// Close закрывает все ресурсы контейнера (для Payment Service не требуется)
func (c *Container) Close() error {
	return nil
}
