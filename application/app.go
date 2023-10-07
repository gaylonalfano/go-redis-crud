package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	// Give this router type a general type (http.Handler), so it's uncoupled from Chi
	router http.Handler
	rdb    *redis.Client
	config Config
}

// Constructor method returns a pointer to our instance of the App type
func New(config Config) *App {
	// Create an instance of our App type and assign to 'app' variable
	app := &App{
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisAddress,
		}),
		config: config,
	}

	// U: Now that we've changed it to (a *App) loadRoutes(),
	// we can just call it directly on the App, since we've already
	// assigned the a.router property to be our router
	app.loadRoutes()

	return app
}

// You define the receiver of this new method using this syntax
// Kinda like the JS 'this' keyword
func (a *App) Start(ctx context.Context) error {
	// Storing 'server' as a pointer, which means we're storing the memory
	// address, NOT as a value!
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}

	// U: Adding this final defer with anon function
	// to ensure it shutdown
	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Println("Failed to close redis", err)
		}
	}()

	fmt.Println("Starting server on port", server.Addr)

	// U: Can't return an error inside this coroutine,
	// but we can use Channel type to communicate across
	// Go routines, e.g., send this error back to the main thread.
	// params: 1 - represents the buffer size of our Channel
	// Channels are buffered (Writer isn't blocked until buffer size
	// is met), or unbuffered (Writer always blocked
	// when writing to a channel). Most times you want unbuffered,
	// but we're using buffered channel here, bc we know only one
	// value will be ever written, and we don't want this Go routine
	// to block if noone is reading from it.
	ch := make(chan error, 1)

	// U: Run our server concurrently using Go coroutines
	// This starts a new thread to run our anon function,
	// and ensures nothing will block
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			// Publish a value onto the Channel
			ch <- fmt.Errorf("Failed to start server: %w", err)
		}
		// Close the Channel and notify using Signals to those listening
		close(ch)
	}()

	// NOTE: Need to listen to TWO channels at once (error and context channels)
	// To do this we use the 'select' keyword, which allows us to block on
	// multiple channels at once. The first channel to have its value to be read,
	// or has its channel closed, will have its case resolved and the code will be
	// able to continue.
	// Now that we use our ctx.Done(), which returns a channel
	select {
	case err = <-ch:
		// Return error case to the caller
		// NOTE: This is why we're using a buffer state for our channel. In the event
		// that this channel isn't called first, then we won't read from this channel
		// again, so we don't wait for our server's Go routine to be deadlocked.
		return err
	case <-ctx.Done():
		// Now we can gracefully shutdown our server
		// Give it 10 seconds to give any inflight requests time to resolve
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		// Close our Redis instance as well using defer cancel()
		defer cancel()

		return server.Shutdown(timeout)
	}

	// NOTE: Basic error channel set up with optional 'open' boolean.
	// Time to setup a receiver for our channel
	// This will capture any value sent in Channel and store into 'err'
	// NOTE: This is blocking until it either receives a value or channel closes (i.e value = nil)
	// err, open := <-ch
	// if !open {
	// 	// Channel was closed
	// 	fmt.Println("Channel was closed")
	// }

	return nil
}
