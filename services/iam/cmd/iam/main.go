package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Импортируем драйвер для database/sql
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	platformlogger "github.com/t4RG3T21/GoBigTech/platform/logger"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/di"
	iampb "github.com/t4RG3T21/GoBigTech/services/iam/v1"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func runMigrations(dbURL string, logger *platformlogger.Logger) error {
	// Парсим URL базы данных для получения имени базы
	// Формат: postgres://user:password@host:port/database?params
	dbName := extractDatabaseName(dbURL)
	if dbName == "" {
		return fmt.Errorf("failed to extract database name from DATABASE_URL")
	}

	// Подключаемся к базе данных postgres (которая всегда существует) для создания нужной БД
	postgresURL := replaceDatabaseName(dbURL, "postgres")
	postgresDB, err := sql.Open("pgx", postgresURL)
	if err != nil {
		return fmt.Errorf("failed to open connection to postgres database: %w", err)
	}
	defer postgresDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем подключение к postgres
	if err := postgresDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres database (make sure PostgreSQL is running and credentials are correct): %w", err)
	}

	// Проверяем, существует ли нужная база данных
	var exists bool
	err = postgresDB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Создаем базу данных, если её нет
	if !exists {
		logger.Info("Database does not exist, creating it", zap.String("database", dbName))

		// Отключаем всех пользователей от создаваемой БД (если она существует)
		_, _ = postgresDB.ExecContext(ctx, fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()", dbName))

		// Создаем базу данных (нельзя использовать параметры в CREATE DATABASE)
		// Экранируем имя базы данных для безопасности
		escapedDBName := fmt.Sprintf(`"%s"`, strings.ReplaceAll(dbName, `"`, `""`))
		createDB := fmt.Sprintf("CREATE DATABASE %s", escapedDBName)

		_, err = postgresDB.ExecContext(ctx, createDB)
		if err != nil {
			// Игнорируем ошибку, если БД уже существует (race condition)
			if !isDatabaseExistsError(err) {
				return fmt.Errorf("failed to create database: %w", err)
			}
			logger.Info("Database already exists (created by another process)", zap.String("database", dbName))
		} else {
			logger.Info("Database created successfully", zap.String("database", dbName))
		}
	}

	// Теперь подключаемся к нужной базе данных
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Проверяем подключение с таймаутом
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database (make sure PostgreSQL is running and credentials are correct): %w", err)
	}

	// Запускаем миграции
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Пробуем разные варианты путей к миграциям
	possiblePaths := []string{
		filepath.Join(workDir, "migrations"),
		filepath.Join(workDir, "..", "migrations"),
		filepath.Join(workDir, "..", "..", "migrations"),
		filepath.Join(workDir, "services", "iam", "migrations"),
		"migrations",
		filepath.Join("..", "migrations"),
	}

	var migrationsDir string
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			files, _ := filepath.Glob(filepath.Join(path, "*.sql"))
			if len(files) > 0 {
				migrationsDir = path
				break
			}
		}
	}

	if migrationsDir == "" {
		return fmt.Errorf("migrations directory not found. Tried paths: %v. Current working directory: %s", possiblePaths, workDir)
	}

	logger.Info("Using migrations directory", zap.String("path", migrationsDir))

	// Применяем миграции
	if err := goose.Up(db, migrationsDir); err != nil {
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
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// 3. Инициализация логгера с уровнем из конфигурации
	logger, err := platformlogger.New("iam-service", cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Устанавливаем логгер в контейнер
	container.SetLogger(logger.Logger)

	logger.Info("Starting IAM Service",
		zap.String("grpc_port", cfg.GRPCPort),
		zap.String("http_port", cfg.HTTPPort),
		zap.String("log_level", cfg.LogLevel),
		zap.String("database_url", maskDatabaseURL(cfg.DatabaseURL)),
		zap.String("redis_address", cfg.RedisAddress),
		zap.Duration("session_ttl", cfg.SessionTTL),
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

	// Получаем gRPC handler через DI контейнер
	authHandler, err := container.AuthHandler(initCtx)
	if err != nil {
		logger.Fatal("Failed to initialize auth handler",
			zap.Error(err),
			zap.String("hint", "Check database and Redis connections"),
		)
	}
	logger.Info("Dependencies initialized successfully")

	// 6. Настройка gRPC сервера
	// Проверяем, свободен ли порт перед созданием listener
	addr := ":" + cfg.GRPCPort
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// Проверяем, является ли ошибка ошибкой занятого порта
		if strings.Contains(err.Error(), "bind") || strings.Contains(err.Error(), "address already in use") {
			logger.Fatal("Port is already in use",
				zap.Error(err),
				zap.String("port", cfg.GRPCPort),
				zap.String("hint", fmt.Sprintf("Another instance of IAM Service might be running on port %s. To find and stop it:\n  Windows: netstat -ano | findstr :%s\n  Then: taskkill /PID <PID> /F\n  Or change GRPC_PORT environment variable to use a different port.", cfg.GRPCPort, cfg.GRPCPort)),
			)
		}
		logger.Fatal("Failed to create listener",
			zap.Error(err),
			zap.String("port", cfg.GRPCPort),
		)
	}
	defer listener.Close() // Закрываем listener при выходе

	grpcServer := grpc.NewServer()
	iampb.RegisterIAMServiceServer(grpcServer, authHandler)

	// Регистрируем gRPC reflection для работы с grpcurl
	reflection.Register(grpcServer)

	// 7. Настройка HTTP Gateway сервера
	ctx := context.Background()

	// Создаем gRPC Gateway mux
	gwMux := runtime.NewServeMux()

	// Регистрируем gateway handlers напрямую к серверу (без проксирования через gRPC клиент)
	// Это более эффективно и избегает проблем с подключением
	if err := iampb.RegisterIAMServiceHandlerServer(ctx, gwMux, authHandler); err != nil {
		logger.Fatal("Failed to register gateway handlers", zap.Error(err))
	}

	// Создаем HTTP сервер с поддержкой HTTP/2
	httpMux := http.NewServeMux()
	httpMux.Handle("/", gwMux)

	// Добавляем CORS middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		httpMux.ServeHTTP(w, r)
	})

	// Используем h2c для поддержки HTTP/2 без TLS
	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	// 8. Запуск gRPC сервера в отдельной горутине
	grpcServerErr := make(chan error, 1)
	go func() {
		logger.Info("gRPC server starting", zap.String("address", listener.Addr().String()))
		if err := grpcServer.Serve(listener); err != nil {
			// Игнорируем ошибку закрытия listener (это нормально при shutdown)
			if err.Error() != "use of closed network connection" {
				grpcServerErr <- err
			}
		}
	}()

	// 9. Запуск HTTP Gateway сервера в отдельной горутине
	httpServerErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP Gateway server starting", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpServerErr <- err
		}
	}()

	// 10. Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-grpcServerErr:
		logger.Fatal("gRPC server error", zap.Error(err))
	case err := <-httpServerErr:
		logger.Fatal("HTTP Gateway server error", zap.Error(err))
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

		// Останавливаем HTTP Gateway сервер
		logger.Info("Shutting down HTTP Gateway server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Warn("Error shutting down HTTP server", zap.Error(err))
		} else {
			logger.Info("HTTP Gateway server stopped")
		}

		// Останавливаем gRPC сервер
		logger.Info("Shutting down gRPC server...")
		// GracefulStop ждет завершения всех активных запросов
		grpcServer.GracefulStop()

		// Закрываем listener
		if err := listener.Close(); err != nil {
			logger.Warn("Error closing listener", zap.Error(err))
		}

		// gRPC connection больше не нужен, так как используем прямой вызов сервера

		// Закрываем ресурсы контейнера (БД, Redis)
		logger.Info("Closing container resources...")
		if err := container.Close(); err != nil {
			logger.Error("Error closing container resources", zap.Error(err))
		} else {
			logger.Info("Container resources closed successfully")
		}
	}

	logger.Info("IAM Service stopped")
}

// extractDatabaseName извлекает имя базы данных из URL
func extractDatabaseName(dbURL string) string {
	// Формат: postgres://user:password@host:port/database?params
	parts := strings.Split(dbURL, "/")
	if len(parts) < 4 {
		return ""
	}
	dbPart := parts[len(parts)-1]
	// Убираем параметры запроса
	if idx := strings.Index(dbPart, "?"); idx != -1 {
		dbPart = dbPart[:idx]
	}
	return dbPart
}

// replaceDatabaseName заменяет имя базы данных в URL
func replaceDatabaseName(dbURL string, newDBName string) string {
	// Формат: postgres://user:password@host:port/database?params
	parts := strings.Split(dbURL, "/")
	if len(parts) < 4 {
		return dbURL
	}
	lastPart := parts[len(parts)-1]
	// Сохраняем параметры запроса
	var params string
	if idx := strings.Index(lastPart, "?"); idx != -1 {
		params = lastPart[idx:]
		lastPart = lastPart[:idx]
	}
	// Заменяем имя базы данных
	parts[len(parts)-1] = newDBName + params
	return strings.Join(parts, "/")
}

// isDatabaseExistsError проверяет, является ли ошибка ошибкой "база данных уже существует"
func isDatabaseExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "SQLSTATE 42P04")
}

// maskDatabaseURL скрывает пароль в URL базы данных для логирования
func maskDatabaseURL(url string) string {
	if len(url) > 0 {
		atPos := -1
		for i := 0; i < len(url); i++ {
			if url[i] == '@' {
				atPos = i
				break
			}
		}
		if atPos > 0 {
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
