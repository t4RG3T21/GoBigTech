package di

import (
	"sync"

	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"

	"github.com/t4RG3T21/GoBigTech/services/notification/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/notification/internal/config"
	"github.com/t4RG3T21/GoBigTech/services/notification/internal/service"
)

// Container - DI контейнер для Notification Service с ленивой загрузкой зависимостей
type Container struct {
	// Конфигурация инициализируется один раз через sync.Once
	configOnce sync.Once
	cfg        *config.Config

	// Логгер
	loggerOnce sync.Once
	logger     *platformlogger.Logger

	// Сервис уведомлений
	serviceOnce sync.Once
	service     *service.NotificationService
	serviceErr  error

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

// Service возвращает сервис уведомлений с ленивой инициализацией
func (c *Container) Service() (*service.NotificationService, error) {
	c.serviceOnce.Do(func() {
		c.service, c.serviceErr = service.NewNotificationService(
			c.cfg.TelegramBotToken,
			c.cfg.TelegramChatID,
			c.logger.Logger,
		)
	})
	return c.service, c.serviceErr
}

// KafkaHandler возвращает Kafka handler с ленивой инициализацией
func (c *Container) KafkaHandler() (*api.KafkaHandler, error) {
	var err error
	c.handlerOnce.Do(func() {
		notificationSvc, svcErr := c.Service()
		if svcErr != nil {
			err = svcErr
			return
		}

		c.handler = api.NewKafkaHandler(
			c.cfg.KafkaBootstrapServers,
			c.cfg.ConsumerGroupID,
			c.cfg.PaymentTopic,
			c.cfg.AssemblyTopic,
			notificationSvc,
			c.logger.Logger,
		)
	})
	return c.handler, err
}

// Close закрывает все ресурсы контейнера
func (c *Container) Close() error {
	if c.handler != nil {
		return c.handler.Close()
	}
	return nil
}
