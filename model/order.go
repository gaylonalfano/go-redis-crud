package model

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags adds a struct tag for JSON type, which allows
// use to encode/decode to JSON using standard libary
// NOTE: Using a combination of timestamps to represent
// the different order statuses. Neat.
type Order struct {
	OrderID     uint64     `json:"order_id"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	LineItems   []LineItem `json:"line_items"`
	CreatedAt   *time.Time `json:"created_at"`
	ShippedAt   *time.Time `json:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type LineItem struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity uint      `json:"quantity"`
	Price    uint      `json:"price"`
}
