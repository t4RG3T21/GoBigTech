package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/models"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByName(ctx context.Context, name string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// PostgresUserRepo реализует UserRepository для PostgreSQL
type PostgresUserRepo struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepo создает новый репозиторий пользователей
func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

// Create создает нового пользователя
func (r *PostgresUserRepo) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO users (id, email, name, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.Email, user.Name, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		// Проверяем на дубликат email
		if isUniqueViolation(err) {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID получает пользователя по ID
func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByName получает пользователя по имени (username)
func (r *PostgresUserRepo) GetByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE name = $1", name).
		Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// Update обновляет информацию о пользователе
func (r *PostgresUserRepo) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET email = $1, name = $2, password_hash = $3, updated_at = $4 WHERE id = $5",
		user.Email, user.Name, user.Password, time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// isUniqueViolation проверяет, является ли ошибка нарушением уникального ограничения
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL возвращает код ошибки 23505 для нарушения уникального ограничения
	// Проверяем текст ошибки на наличие ключевых слов
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "23505")
}
