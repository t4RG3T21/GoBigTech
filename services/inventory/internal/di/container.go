package di

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/config"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/inventory/internal/service"
)

// Container - DI контейнер для Inventory Service с ленивой загрузкой зависимостей
type Container struct {
	// Конфигурация инициализируется один раз через sync.Once
	configOnce sync.Once
	cfg        *config.Config

	// Зависимости с ленивой инициализацией
	mongoClientOnce sync.Once
	mongoClient     *mongo.Client
	mongoClientErr  error

	collectionOnce sync.Once
	collection     *mongo.Collection
	collectionErr  error

	repositoryOnce sync.Once
	repository     *repository.MongoInventoryRepo
	repositoryErr  error

	serviceOnce sync.Once
	service     *service.InventoryService
	serviceErr  error
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

// MongoClient возвращает клиент MongoDB с ленивой инициализацией
func (c *Container) MongoClient(ctx context.Context) (*mongo.Client, error) {
	c.mongoClientOnce.Do(func() {
		cfg := c.Config()

		// Создаем контекст с таймаутом для подключения
		connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(cfg.MongoDBURL))
		if err != nil {
			c.mongoClientErr = fmt.Errorf("failed to connect to MongoDB: %w", err)
			return
		}

		// Проверяем подключение
		if err := client.Ping(ctx, nil); err != nil {
			client.Disconnect(ctx)
			c.mongoClientErr = fmt.Errorf("failed to ping MongoDB: %w", err)
			return
		}

		c.mongoClient = client
	})

	return c.mongoClient, c.mongoClientErr
}

// Collection возвращает коллекцию MongoDB с ленивой инициализацией
func (c *Container) Collection(ctx context.Context) (*mongo.Collection, error) {
	c.collectionOnce.Do(func() {
		client, err := c.MongoClient(ctx)
		if err != nil {
			c.collectionErr = fmt.Errorf("failed to get MongoDB client: %w", err)
			return
		}

		cfg := c.Config()
		c.collection = client.Database(cfg.DatabaseName).Collection(cfg.CollectionName)
	})

	return c.collection, c.collectionErr
}

// Repository возвращает репозиторий с ленивой инициализацией
func (c *Container) Repository(ctx context.Context) (*repository.MongoInventoryRepo, error) {
	c.repositoryOnce.Do(func() {
		collection, err := c.Collection(ctx)
		if err != nil {
			c.repositoryErr = fmt.Errorf("failed to get collection: %w", err)
			return
		}

		c.repository = repository.NewMongoInventoryRepo(collection)
	})

	return c.repository, c.repositoryErr
}

// Service возвращает сервис инвентаря с ленивой инициализацией
func (c *Container) Service(ctx context.Context) (*service.InventoryService, error) {
	c.serviceOnce.Do(func() {
		repo, err := c.Repository(ctx)
		if err != nil {
			c.serviceErr = fmt.Errorf("failed to get repository: %w", err)
			return
		}

		c.service = service.NewInventoryService(repo)
	})

	return c.service, c.serviceErr
}

// Close закрывает все ресурсы контейнера
func (c *Container) Close() error {
	if c.mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.mongoClient.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB: %w", err)
		}
	}
	return nil
}
