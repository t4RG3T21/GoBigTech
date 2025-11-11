package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Импортируем драйвер для database/sql
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderapi "github.com/t4RG3T21/GoBigTech/services/order/api"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/api"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/repository"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/service"

	inventorypb "github.com/t4RG3T21/GoBigTech/services/inventory/v1"
	paymentpb "github.com/t4RG3T21/GoBigTech/services/payment/v1"
)

func runMigrations(dbURL string) error {
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
	// Получаем рабочую директорию
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

	log.Printf("Using migrations directory: %s", migrationsDir)

	// Проверяем, существуют ли таблицы (если созданы вручную через DBeaver)
	var ordersExists, orderItemsExists bool
	row := db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'orders')")
	if err := row.Scan(&ordersExists); err != nil {
		log.Printf("Warning: Could not check if orders table exists: %v", err)
	}

	row = db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'order_items')")
	if err := row.Scan(&orderItemsExists); err != nil {
		log.Printf("Warning: Could not check if order_items table exists: %v", err)
	}

	// Если таблицы уже существуют, помечаем миграцию как примененную
	if ordersExists && orderItemsExists {
		log.Printf("Tables 'orders' and 'order_items' already exist (created manually)")

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
				log.Printf("Warning: Could not set migration version: %v", err)
			} else {
				log.Printf("Migration version set to 1 (tables already exist)")
			}
		} else if gooseTableExists {
			currentVersion, err := goose.GetDBVersion(db)
			if err == nil {
				log.Printf("Current migration version: %d", currentVersion)
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
			log.Printf("Warning: Migration error, but tables exist. Continuing...")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func main() {
	// 1. Запуск миграций
	// Используем переменную окружения или значение по умолчанию
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Используем 127.0.0.1 вместо localhost, чтобы избежать проблем с IPv6
		dbURL = "postgres://postgres:password@127.0.0.1:5432/order_service?sslmode=disable"
	}

	log.Printf("Connecting to database: postgres://postgres:***@127.0.0.1:5432/order_service")
	if err := runMigrations(dbURL); err != nil {
		log.Printf("Error details: %v", err)
		log.Fatalf(`
Failed to connect to PostgreSQL. Please check:

1. Is PostgreSQL running?
   - Docker: docker-compose up -d postgres (from services directory)
   - Local: Check if PostgreSQL service is running

2. Are the credentials correct?
   - Default Docker: postgres/password
   - Set custom: export DATABASE_URL="postgres://user:pass@127.0.0.1:5432/order_service?sslmode=disable"

3. Is the database created?
   - Docker: Created automatically
   - Local: CREATE DATABASE order_service;

4. Check connection:
   - Docker: docker-compose exec postgres psql -U postgres -d order_service
   - Local: psql -U postgres -d order_service

Original error: %v`, err)
	}
	log.Println("Migrations completed successfully")

	// 2. Подключение к PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Инициализация репозитория
	orderRepo := repository.NewPostgresOrderRepo(db)

	// 4. Подключение к внешним сервисам
	connInv, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to inventory: %v", err)
	}
	defer connInv.Close()

	connPay, err := grpc.NewClient("127.0.0.1:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to payment: %v", err)
	}
	defer connPay.Close()

	// 5. Создание адаптеров
	invAdapter := &service.InventoryClientAdapter{
		Client: inventorypb.NewInventoryServiceClient(connInv),
	}
	payAdapter := &service.PaymentClientAdapter{
		Client: paymentpb.NewPaymentServiceClient(connPay),
	}

	// 6. Создание сервиса
	orderService := service.NewOrderService(orderRepo, invAdapter, payAdapter)

	// 7. Создание HTTP обработчиков
	orderHandler := api.NewOrderHandler(orderService)

	// 8. Настройка маршрутов
	r := chi.NewRouter()

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "order",
		})
	})

	// API routes
	orderapi.HandlerFromMux(orderHandler, r)

	log.Println("Order service (with PostgreSQL) starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
