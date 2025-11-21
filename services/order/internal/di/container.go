package di

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/t4RG3T21/GoBigTech/services/order/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/config"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/service"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

// Container - DI контейнер для Order Service с ленивой загрузкой зависимостей
type Container struct {
	// Конфигурация инициализируется один раз через sync.Once
	configOnce sync.Once
	cfg        *config.Config

	// Зависимости с ленивой инициализацией
	dbOnce            sync.Once
	db                *pgxpool.Pool
	dbErr             error

	orderRepoOnce     sync.Once
	orderRepo         repository.OrderRepository
	orderRepoErr      error

	inventoryClientOnce sync.Once
	inventoryClient     service.InventoryClient
	inventoryClientErr  error

	paymentClientOnce  sync.Once
	paymentClient      service.PaymentClient
	paymentClientErr   error

	orderServiceOnce   sync.Once
	orderService       *service.OrderService
	orderServiceErr    error

	orderHandlerOnce   sync.Once
	orderHandler       *api.OrderHandler
	orderHandlerErr    error

	// gRPC соединения для клиентов
	inventoryConn      *grpc.ClientConn
	paymentConn        *grpc.ClientConn
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

// Database возвращает пул подключений к PostgreSQL с ленивой инициализацией
func (c *Container) Database(ctx context.Context) (*pgxpool.Pool, error) {
	c.dbOnce.Do(func() {
		cfg := c.Config()
		
		// Создаем контекст с таймаутом для подключения
		connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		pool, err := pgxpool.New(connectCtx, cfg.DatabaseURL)
		if err != nil {
			c.dbErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Проверяем подключение
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			c.dbErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		c.db = pool
	})

	return c.db, c.dbErr
}

// OrderRepository возвращает репозиторий заказов с ленивой инициализацией
func (c *Container) OrderRepository(ctx context.Context) (repository.OrderRepository, error) {
	c.orderRepoOnce.Do(func() {
		db, err := c.Database(ctx)
		if err != nil {
			c.orderRepoErr = fmt.Errorf("failed to get database: %w", err)
			return
		}

		c.orderRepo = repository.NewPostgresOrderRepo(db)
	})

	return c.orderRepo, c.orderRepoErr
}

// InventoryClient возвращает клиент для Inventory Service с ленивой инициализацией
func (c *Container) InventoryClient() (service.InventoryClient, error) {
	c.inventoryClientOnce.Do(func() {
		cfg := c.Config()

		// Создаем gRPC соединение
		conn, err := grpc.NewClient(
			cfg.InventoryGRPCAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			c.inventoryClientErr = fmt.Errorf("failed to connect to inventory service: %w", err)
			return
		}

		c.inventoryConn = conn

		// Создаем адаптер
		client := inventorypb.NewInventoryServiceClient(conn)
		c.inventoryClient = &service.InventoryClientAdapter{
			Client: client,
		}
	})

	return c.inventoryClient, c.inventoryClientErr
}

// PaymentClient возвращает клиент для Payment Service с ленивой инициализацией
func (c *Container) PaymentClient() (service.PaymentClient, error) {
	c.paymentClientOnce.Do(func() {
		cfg := c.Config()

		// Создаем gRPC соединение
		conn, err := grpc.NewClient(
			cfg.PaymentGRPCAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			c.paymentClientErr = fmt.Errorf("failed to connect to payment service: %w", err)
			return
		}

		c.paymentConn = conn

		// Создаем адаптер
		client := paymentpb.NewPaymentServiceClient(conn)
		c.paymentClient = &service.PaymentClientAdapter{
			Client: client,
		}
	})

	return c.paymentClient, c.paymentClientErr
}

// OrderService возвращает сервис заказов с ленивой инициализацией
func (c *Container) OrderService(ctx context.Context) (*service.OrderService, error) {
	c.orderServiceOnce.Do(func() {
		// Получаем все зависимости
		repo, err := c.OrderRepository(ctx)
		if err != nil {
			c.orderServiceErr = fmt.Errorf("failed to get order repository: %w", err)
			return
		}

		invClient, err := c.InventoryClient()
		if err != nil {
			c.orderServiceErr = fmt.Errorf("failed to get inventory client: %w", err)
			return
		}

		payClient, err := c.PaymentClient()
		if err != nil {
			c.orderServiceErr = fmt.Errorf("failed to get payment client: %w", err)
			return
		}

		// Создаем сервис
		c.orderService = service.NewOrderService(repo, invClient, payClient)
	})

	return c.orderService, c.orderServiceErr
}

// OrderHandler возвращает HTTP обработчик заказов с ленивой инициализацией
func (c *Container) OrderHandler(ctx context.Context) (*api.OrderHandler, error) {
	c.orderHandlerOnce.Do(func() {
		orderService, err := c.OrderService(ctx)
		if err != nil {
			c.orderHandlerErr = fmt.Errorf("failed to get order service: %w", err)
			return
		}

		c.orderHandler = api.NewOrderHandler(orderService)
	})

	return c.orderHandler, c.orderHandlerErr
}

// Close закрывает все ресурсы контейнера (база данных, gRPC соединения)
func (c *Container) Close() error {
	var errs []error

	if c.db != nil {
		c.db.Close()
	}

	if c.inventoryConn != nil {
		if err := c.inventoryConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close inventory connection: %w", err))
		}
	}

	if c.paymentConn != nil {
		if err := c.paymentConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close payment connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing container: %v", errs)
	}

	return nil
}

