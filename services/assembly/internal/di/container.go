package di

import (
	"sync"

	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"

	"github.com/t4RG3T21/GoBigTech/services/assembly/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/assembly/internal/config"
	"github.com/t4RG3T21/GoBigTech/services/assembly/internal/service"
)

// Container - DI контейнер для Assembly Service с ленивой загрузкой зависимостей
type Container struct {
	// Конфигурация инициализируется один раз через sync.Once
	configOnce sync.Once
	cfg        *config.Config

	// Логгер
	loggerOnce sync.Once
	logger     *platformlogger.Logger

	// Сервис сборки
	serviceOnce sync.Once
	service     *service.AssemblyService

	// Kafka handler
	handlerOnce sync.Once
	handler     *api.KafkaHandler
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

// Logger возвращает логгер (должен быть установлен через SetLogger)
func (c *Container) Logger() *platformlogger.Logger {
	return c.logger
}

// SetLogger устанавливает логгер
func (c *Container) SetLogger(logger *platformlogger.Logger) {
	c.loggerOnce.Do(func() {
		c.logger = logger
	})
}

// Service возвращает сервис сборки с ленивой инициализацией
func (c *Container) Service() *service.AssemblyService {
	c.serviceOnce.Do(func() {
		c.service = service.NewAssemblyService(c.logger.Logger)
	})
	return c.service
}

// KafkaHandler возвращает Kafka handler с ленивой инициализацией
func (c *Container) KafkaHandler() *api.KafkaHandler {
	c.handlerOnce.Do(func() {
		assemblySvc := c.Service()
		c.handler = api.NewKafkaHandler(
			c.cfg.KafkaBootstrapServers,
			c.cfg.ConsumerGroupID,
			c.cfg.PaymentTopic,
			c.cfg.AssemblyTopic,
			assemblySvc,
			c.logger.Logger,
		)
	})
	return c.handler
}

// Close закрывает все ресурсы контейнера
func (c *Container) Close() error {
	if c.handler != nil {
		return c.handler.Close()
	}
	return nil
}
