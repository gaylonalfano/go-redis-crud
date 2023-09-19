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

	// U: Atomic transaction that uses a new pipeline client
	// that wraps queued commands in Redis' MULTI/EXEC. This will
	// replace our old/original r.Client
	// REF: https://youtu.be/qCv-q37qjZU?t=822
	txn := r.Client.TxPipeline()

	// Set() overwrites. SetNX() overwrites if not exists.
	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to set: %w", err)
	}

	// NOTE: For pagination, we don't want to fetch all orders at once, so
	// we're adding a Set (Q: Is this a Redis-specific thing?) that only
	// holds the order IDs for faster FindAll(). However, to keep the db
	// and the set in sync, we use an atomic transaction that will fail
	// if either part fails (like Solana txs). Prevents partial state.
	// NOTE: The set is simply something like this:
	// "orders": "id1, id2, id3, ..."
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to add to orders set: %w", err)
	}

	// U: No buffered Pipeline commands will execute and send to
	// the Redis server until we commit them
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec: %w", err)
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

	// U: Using atomic transaction pipeline client instead for pagination
	// to keep 'orders' and the orders set in sync.
	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to delete order: %w", err)
	}

	// U: Remove the id from the orders set
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to remove from orders set: %w", err)
	}

	// Send the queued pipeline command buffers to redis server
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec: %w", err)
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

// In order to support pagination, rather than fetching all at once,
// we create a new type with a couple properties we can use to help
type FindAllPage struct {
	Size   uint64 // aka Count
	Offset uint64 // aka Cursor
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	// Let's get all the IDs within the specified page range
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	// Now let's extract each piece from the res.Result()
	// TODO: Using a set returns unordered values. There is an OrderedSet Redis option,
	// but that could be an extension exercise
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("Failed to get order ids from set: %w", err)
	}

	// Check the 'keys' size. If keys are empty, then return empty list
	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
			Cursor: 0,
		}, nil
	}

	// We only have the IDs. Now time to get all the full values for each key
	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("Failed to get order values from keys: %w", err)
	}

	// Unwrap these orders values into an orders slice (for pagination)
	orders := make([]model.Order, len(xs))

	// Iterate over each element and case each to a string
	for i, x := range xs {
		x := x.(string)
		var order model.Order

		// Then, Unmarshal (decode) string (x) into an Order struct
		// which we'll store in our orders slice at the current index (i)
		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("Failed to decode order json: %w", err)
		}

		orders[i] = order
	}

	// Return our FindResult with orders and the next cursor value
	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
