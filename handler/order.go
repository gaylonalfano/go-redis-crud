package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/gaylonalfano/go-turso-htmx-orders/model"
	"github.com/gaylonalfano/go-turso-htmx-orders/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create an order")
	// 'body' has anonymous type and declared inline. 'body' will
	// represent the expected POST data from client
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		// Send bad status code if fails, since we'd send bad input data
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Construct our model.Order so we can insert it
	now := time.Now().UTC() // time.Time
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now, // memory address only (*time.Time)
	}

	err := o.Repo.Insert(r.Context(), order)
	if err != nil {
		fmt.Println("Failed to insert:", err)
		// Send 500 code since something broke on our end
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return our generated model.Order to the Client
	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Failed to encode:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated) // 201
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("List all orders")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order by ID")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order by ID")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order by ID")
}
