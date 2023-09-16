package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
	// Give this router type a general type, so it's uncoupled from Chi
	router http.Handler
}

// Constructor method returns a pointer to our instance of the App type
func New() *App {
	// Create an instance of our App type and assign to 'app' variable
	app := &App{
		router: loadRoutes(),
	}

	return app
}

// You define the receiver of this new method using this syntax
// Kinda like the JS 'this' keyword
func (a *App) Start(ctx context.Context) error {
	// Storing 'server' as a pointer, which means we're storing the memory
	// address, NOT as a value!
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
	}

	return nil
}
