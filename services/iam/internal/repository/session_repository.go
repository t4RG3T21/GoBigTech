package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/t4RG3T21/GoBigTech/services/iam/internal/models"
)

// SessionRepository определяет интерфейс для работы с сессиями
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (*models.Session, error)
	Delete(ctx context.Context, sessionID string) error
	Refresh(ctx context.Context, sessionID string, ttl time.Duration) error
}

// RedisSessionRepo реализует SessionRepository для Redis
type RedisSessionRepo struct {
	client *redis.Client
}

// NewRedisSessionRepo создает новый репозиторий сессий
func NewRedisSessionRepo(client *redis.Client) *RedisSessionRepo {
	return &RedisSessionRepo{client: client}
}

// Create создает новую сессию в Redis
func (r *RedisSessionRepo) Create(ctx context.Context, session *models.Session, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s", session.ID)
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session in redis: %w", err)
	}

	return nil
}

// Get получает сессию из Redis
func (r *RedisSessionRepo) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Проверяем, не истекла ли сессия
	if time.Now().After(session.ExpiresAt) {
		// Удаляем истекшую сессию
		_ = r.Delete(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// Delete удаляет сессию из Redis
func (r *RedisSessionRepo) Delete(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from redis: %w", err)
	}
	return nil
}

// Refresh обновляет TTL сессии
func (r *RedisSessionRepo) Refresh(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to refresh session in redis: %w", err)
	}
	return nil
}
