package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
)

type PostgresOrderRepo struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepo(db *pgxpool.Pool) *PostgresOrderRepo {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Вставляем заказ
	_, err = tx.Exec(ctx,
		"INSERT INTO orders (id, user_id, status, total) VALUES ($1, $2, $3, $4)",
		order.ID, order.UserID, order.Status, order.Total)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Вставляем элементы заказа
	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			"INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)",
			order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresOrderRepo) GetByID(ctx context.Context, id string) (*models.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var order models.Order
	err = tx.QueryRow(ctx,
		"SELECT id, user_id, status, total FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.UserID, &order.Status, &order.Total)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	rows, err := tx.Query(ctx,
		"SELECT product_id, quantity, price FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	order.Items = items

	return &order, nil
}

func (r *PostgresOrderRepo) Update(ctx context.Context, order *models.Order) error {
	_, err := r.db.Exec(ctx,
		"UPDATE orders SET status = $1, total = $2, updated_at = $3 WHERE id = $4",
		order.Status, order.Total, time.Now(), order.ID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}
