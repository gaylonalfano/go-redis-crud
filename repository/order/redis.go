package order

// NOTE: This whole repo/order/ package is a layer of abstraction,
// so that we can easily swap out our db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gaylonalfano/go-turso-htmx-orders/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func generateOrderIDKey(id uint64) string {
	return fmt.Sprintln("order:%d", id)
}

// NOTE: Redis is a k:v store, stored as string, so we're using the JSON
// encode (Marshal) / decode (UnMarshal) for this (look at our Order type).
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order) // []byte
	if err != nil {
		return fmt.Errorf("Failed to encode order: %w", err)
	}

	key := generateOrderIDKey(order.OrderID)

	// Set() overwrites. SetNX() overwrites if not exists.
	res := r.Client.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		return fmt.Errorf("Failed to set: %w", err)
	}

	return nil
}

// Create a custom error (Redis does have a r.Nil() error)
var ErrNotExist = errors.New("Order does not exist")

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := generateOrderIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	// Check whether the error is a Redis error, so we can then
	// return our custom error.
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("Failed to get order: %w", err)
	}

	// Decode the JSON into a proper Order
	var order model.Order
	// NOTE: & seems to specify the 'pointer' of the value
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("Failed to decode order json: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := generateOrderIDKey(id)

	err := r.Client.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("Failed to delete order: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Failed to encode order: %w", err)
	}

	key := generateOrderIDKey(order.OrderID)

	// SetXX() only sets/updates value if already exists
	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		// It's an error but not a redis.Nil error type
		return fmt.Errorf("Failed to update order: %w", err)
	}

	return nil
}
