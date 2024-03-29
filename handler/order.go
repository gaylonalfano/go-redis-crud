package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	// "text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaylonalfano/go-redis-crud/model"
	"github.com/gaylonalfano/go-redis-crud/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
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
		OrderID:    rand.Uint64(), // Not for production!
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now, // memory address only (*time.Time)
	}

	err := h.Repo.Insert(r.Context(), order)
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

func (h *Order) List(w http.ResponseWriter, r *http.Request) {
	// Users will pass in a query param for cursor or page number (pagination)
	cursorStr := r.URL.Query().Get("cursor")
	// If nothing passed, then set to 0
	if cursorStr == "" {
		cursorStr = "0"
	}
	// Parse cursor into a uint64
	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	// FIXME: Cursor is always zero, so never get next cursor position
	fmt.Println("strconv.ParseUint() cursor:", cursor) // 0!
	if err != nil {
		fmt.Println("Bad cursor", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Call our Repo's FindAll()
	const size = 50
	res, err := h.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	fmt.Println("FindResult:", res)
	if err != nil {
		fmt.Println("Failed to find all:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Craft our response with an anonymous struct
	// Using omitempty if Next == 0, i.e. no more pages
	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor
	fmt.Println("response struct:", response)

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)

	// U: Experimenting with Go Templates + HTMX
	// t := template.Must(template.ParseFiles("index.html"))
	t := template.Must(template.New("index.html").Parse("index.html"))

	// Q: What data to pass? Encoded JSON?
	// Q: Should I use text/template package instead of html/template?
	// t.ExecuteTemplate(w, "order-list-element", response.Items) // json output
	// t.ExecuteTemplate(w, "order-list-element", data)
	t.Execute(w, data)

}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order by ID")
	idParam := chi.URLParam(r, "id")

	// Convert to uint64
	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, err := h.Repo.FindByID(r.Context(), orderID)
	// Check whether err is our custom error
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Encode the order type directly into the ResponseWriter
	// Q: Is json.NewEncoder(w).Encode(o) same as json.Marshal(r)?
	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("Failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order by ID")
	// 'body' to represent PUT data from client
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Pull out the Order ID
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Retrieve existing order
	currentOrder, err := h.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only allow updating Order if certain conditions met
	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		if currentOrder.ShippedAt != nil {
			// TODO: Send by custom error messages to client
			fmt.Println("Failed to update status to:", shippedStatus)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		currentOrder.ShippedAt = &now
	case completedStatus:
		if currentOrder.CompletedAt != nil || currentOrder.ShippedAt == nil {
			fmt.Println("Failed to update status to:", completedStatus)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		currentOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.Update(r.Context(), currentOrder)
	if err != nil {
		fmt.Println("Failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If all is well, send it back to client encoded as JSON
	if err := json.NewEncoder(w).Encode(currentOrder); err != nil {
		fmt.Println("Failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order by ID")
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
