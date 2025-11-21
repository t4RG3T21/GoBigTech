package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // Импортируем драйвер для database/sql
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	orderapi "github.com/t4RG3T21/GoBigTech/services/order/api"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/di"
	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
)

func runMigrations(dbURL string, logger *platformlogger.Logger) error {
	// Используем pgx драйвер через stdlib для совместимости с database/sql
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Проверяем подключение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database (make sure PostgreSQL is running and credentials are correct): %w", err)
	}

	// Запускаем миграции
	// Определяем путь к миграциям относительно корня модуля
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Пробуем разные варианты путей к миграциям
	possiblePaths := []string{
		filepath.Join(workDir, "migrations"),                      // Если запускаем из services/order
		filepath.Join(workDir, "..", "migrations"),                // Если запускаем из services/order/cmd/order
		filepath.Join(workDir, "..", "..", "migrations"),          // Если запускаем из cmd/order
		filepath.Join(workDir, "services", "order", "migrations"), // Если запускаем из корня проекта
		"migrations",                      // Относительный путь
		filepath.Join("..", "migrations"), // Относительный путь вверх
	}

	var migrationsDir string
	for _, path := range possiblePaths {
		// Проверяем, что путь существует и это директория
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Проверяем, что в директории есть SQL файлы
			files, _ := filepath.Glob(filepath.Join(path, "*.sql"))
			if len(files) > 0 {
				migrationsDir = path
				break
			}
		}
	}

	if migrationsDir == "" {
		return fmt.Errorf("migrations directory not found. Tried paths: %v. Current working directory: %s. Please ensure migrations/001_create_orders_table.sql exists", possiblePaths, workDir)
	}

	logger.Info("Using migrations directory", zap.String("path", migrationsDir))

	// Проверяем, существуют ли таблицы (если созданы вручную через DBeaver)
	var ordersExists, orderItemsExists bool
	row := db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'orders')")
	if err := row.Scan(&ordersExists); err != nil {
		logger.Warn("Could not check if orders table exists", zap.Error(err))
	}

	row = db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'order_items')")
	if err := row.Scan(&orderItemsExists); err != nil {
		logger.Warn("Could not check if order_items table exists", zap.Error(err))
	}

	// Если таблицы уже существуют, помечаем миграцию как примененную
	if ordersExists && orderItemsExists {
		logger.Info("Tables 'orders' and 'order_items' already exist (created manually)")

		// Проверяем, есть ли таблица версий goose
		var gooseTableExists bool
		row = db.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'goose_db_version')")
		if err := row.Scan(&gooseTableExists); err == nil && !gooseTableExists {
			// Создаем таблицу версий goose и помечаем миграцию как примененную
			goose.SetDialect("postgres")
			// Создаем таблицу версий вручную и устанавливаем версию 1
			_, err = db.ExecContext(ctx, `
				CREATE TABLE IF NOT EXISTS goose_db_version (
					id SERIAL PRIMARY KEY,
					version_id BIGINT NOT NULL,
					is_applied BOOLEAN NOT NULL,
					tstamp TIMESTAMP DEFAULT NOW()
				);
				INSERT INTO goose_db_version (version_id, is_applied) 
				SELECT 1, true 
				WHERE NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version_id = 1);
			`)
			if err != nil {
				logger.Warn("Could not set migration version", zap.Error(err))
			} else {
				logger.Info("Migration version set to 1 (tables already exist)")
			}
		} else if gooseTableExists {
			currentVersion, err := goose.GetDBVersion(db)
			if err == nil {
				logger.Info("Current migration version", zap.Int64("version", currentVersion))
			}
		}
		return nil
	}

	// Применяем миграции (goose автоматически пропустит уже примененные)
	if err := goose.Up(db, migrationsDir); err != nil {
		// Проверяем, может быть таблицы все-таки создались
		row = db.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'orders')")
		if err := row.Scan(&ordersExists); err == nil && ordersExists {
			logger.Warn("Migration error, but tables exist. Continuing...")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func main() {
	// 1. Создание DI контейнера
	container := di.NewContainer()

	// 2. Получение конфигурации
	cfg := container.Config()
	if err := cfg.Validate(); err != nil {
		// Используем стандартный вывод для ошибки конфигурации
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// 3. Инициализация логгера с уровнем из конфигурации
	logger, err := platformlogger.New("order-service", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() // Обеспечиваем запись всех логов при завершении

	logger.Info("Starting Order Service",
		zap.String("http_port", cfg.HTTPPort),
		zap.String("log_level", cfg.LogLevel),
		zap.String("database_url", maskDatabaseURL(cfg.DatabaseURL)),
		zap.String("inventory_grpc", cfg.InventoryGRPCAddress),
		zap.String("payment_grpc", cfg.PaymentGRPCAddress),
	)

	// 4. Запуск миграций базы данных
	logger.Info("Running database migrations")
	if err := runMigrations(cfg.DatabaseURL, logger); err != nil {
		logger.Fatal("Failed to run migrations",
			zap.Error(err),
			zap.String("hint", "Make sure PostgreSQL is running and DATABASE_URL is correct"),
		)
	}
	logger.Info("Migrations completed successfully")

	// 5. Инициализация зависимостей через DI контейнер
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	// Получаем handler через DI контейнер (он инициализирует всю цепочку зависимостей)
	orderHandler, err := container.OrderHandler(initCtx)
	if err != nil {
		logger.Fatal("Failed to initialize order handler",
			zap.Error(err),
			zap.String("hint", "Check database connection and gRPC services availability"),
		)
	}
	logger.Info("Dependencies initialized successfully")

	// 6. Настройка HTTP роутера
	r := chi.NewRouter()

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"service":   "order",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API routes
	orderapi.HandlerFromMux(orderHandler, r)

	// 7. Настройка HTTP сервера
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Запуск сервера в отдельной горутине
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// 9. Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Fatal("Server error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Создаем контекст с таймаутом для graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Останавливаем HTTP сервер
		logger.Info("Shutting down HTTP server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during server shutdown", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped gracefully")
		}

		// Закрываем ресурсы контейнера (БД, gRPC соединения)
		logger.Info("Closing container resources...")
		if err := container.Close(); err != nil {
			logger.Error("Error closing container resources", zap.Error(err))
		} else {
			logger.Info("Container resources closed successfully")
		}
	}

	logger.Info("Order Service stopped")
}

// maskDatabaseURL скрывает пароль в URL базы данных для логирования
func maskDatabaseURL(url string) string {
	// Простая маскировка: заменяем пароль на ***
	// Формат: postgres://user:password@host:port/db
	if len(url) > 0 {
		// Ищем позицию @ после пароля
		atPos := -1
		for i := 0; i < len(url); i++ {
			if url[i] == '@' {
				atPos = i
				break
			}
		}
		if atPos > 0 {
			// Ищем двоеточие перед паролем
			colonPos := -1
			for i := 0; i < atPos; i++ {
				if url[i] == ':' {
					colonPos = i
					break
				}
			}
			if colonPos > 0 {
				return url[:colonPos+1] + "***" + url[atPos:]
			}
		}
	}
	return url
}
