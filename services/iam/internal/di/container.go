package di

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/iam/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/config"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/service"
)

// Container - DI контейнер для IAM Service с ленивой загрузкой зависимостей
type Container struct {
	// Конфигурация инициализируется один раз через sync.Once
	configOnce sync.Once
	cfg        *config.Config

	// Зависимости с ленивой инициализацией
	dbOnce sync.Once
	db     *pgxpool.Pool
	dbErr  error

	redisOnce   sync.Once
	redisClient *redis.Client
	redisErr    error

	userRepoOnce sync.Once
	userRepo     repository.UserRepository
	userRepoErr  error

	sessionRepoOnce sync.Once
	sessionRepo     repository.SessionRepository
	sessionRepoErr  error

	authServiceOnce sync.Once
	authService     *service.AuthService
	authServiceErr  error

	authHandlerOnce sync.Once
	authHandler     *api.AuthHandler
	authHandlerErr  error

	// Логгер
	logger *zap.Logger
}

// NewContainer создает новый DI контейнер
func NewContainer() *Container {
	return &Container{}
}

// SetLogger устанавливает логгер
func (c *Container) SetLogger(logger *zap.Logger) {
	c.logger = logger
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

// Redis возвращает клиент Redis с ленивой инициализацией
func (c *Container) Redis(ctx context.Context) (*redis.Client, error) {
	c.redisOnce.Do(func() {
		cfg := c.Config()

		c.redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddress,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})

		// Проверяем подключение
		if err := c.redisClient.Ping(ctx).Err(); err != nil {
			c.redisErr = fmt.Errorf("failed to ping redis: %w", err)
			return
		}
	})

	return c.redisClient, c.redisErr
}

// UserRepository возвращает репозиторий пользователей с ленивой инициализацией
func (c *Container) UserRepository(ctx context.Context) (repository.UserRepository, error) {
	c.userRepoOnce.Do(func() {
		db, err := c.Database(ctx)
		if err != nil {
			c.userRepoErr = fmt.Errorf("failed to get database: %w", err)
			return
		}

		c.userRepo = repository.NewPostgresUserRepo(db)
	})

	return c.userRepo, c.userRepoErr
}

// SessionRepository возвращает репозиторий сессий с ленивой инициализацией
func (c *Container) SessionRepository(ctx context.Context) (repository.SessionRepository, error) {
	c.sessionRepoOnce.Do(func() {
		redisClient, err := c.Redis(ctx)
		if err != nil {
			c.sessionRepoErr = fmt.Errorf("failed to get redis client: %w", err)
			return
		}

		c.sessionRepo = repository.NewRedisSessionRepo(redisClient)
	})

	return c.sessionRepo, c.sessionRepoErr
}

// AuthService возвращает сервис аутентификации с ленивой инициализацией
func (c *Container) AuthService(ctx context.Context) (*service.AuthService, error) {
	c.authServiceOnce.Do(func() {
		if c.logger == nil {
			c.authServiceErr = fmt.Errorf("logger not set")
			return
		}

		userRepo, err := c.UserRepository(ctx)
		if err != nil {
			c.authServiceErr = fmt.Errorf("failed to get user repository: %w", err)
			return
		}

		sessionRepo, err := c.SessionRepository(ctx)
		if err != nil {
			c.authServiceErr = fmt.Errorf("failed to get session repository: %w", err)
			return
		}

		cfg := c.Config()
		c.authService = service.NewAuthService(userRepo, sessionRepo, cfg.SessionTTL, c.logger)
	})

	return c.authService, c.authServiceErr
}

// AuthHandler возвращает gRPC обработчик с ленивой инициализацией
func (c *Container) AuthHandler(ctx context.Context) (*api.AuthHandler, error) {
	c.authHandlerOnce.Do(func() {
		if c.logger == nil {
			c.authHandlerErr = fmt.Errorf("logger not set")
			return
		}

		authService, err := c.AuthService(ctx)
		if err != nil {
			c.authHandlerErr = fmt.Errorf("failed to get auth service: %w", err)
			return
		}

		c.authHandler = api.NewAuthHandler(authService, c.logger)
	})

	return c.authHandler, c.authHandlerErr
}

// Close закрывает все ресурсы контейнера (база данных, Redis)
func (c *Container) Close() error {
	var errs []error

	if c.db != nil {
		c.db.Close()
		// Даем время на закрытие соединений
		time.Sleep(100 * time.Millisecond)
	}

	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing container: %v", errs)
	}

	return nil
}
