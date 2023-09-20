package application

import (
	"net/http"

	"github.com/gaylonalfano/go-turso-htmx-orders/handler"
	"github.com/gaylonalfano/go-turso-htmx-orders/repository/order"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// U: We need to provide our handlers with an instance of our RedisRepo,
// This means we can make loadRoutes be part of the App struct itself,
// to easily access App properties
func (a *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// func(){} is an anonymous function syntax
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create/setup a subrouter for the /orders path
	// NOTE: This is a short-hand for Mount()
	router.Route("/orders", a.loadOrderRoutes)

	// U: Instead of returning a *chi.Mux router, we just
	// update/assign our App's router property to this router
	a.router = router
}

func (a *App) loadOrderRoutes(router chi.Router) {
	// Use '&' to take the memory address of the instance
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: a.rdb,
		},
	}

	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.GetByID)
	router.Put("/{id}", orderHandler.UpdateByID)
	router.Delete("/{id}", orderHandler.DeleteByID)
}
