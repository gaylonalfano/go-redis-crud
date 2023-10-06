package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gaylonalfano/go-turso-htmx-orders/application"
)

// NOTE:
// - Get Docker going: docker run -p 6379:6379 redis:latest
// - Get our server going: go run main.go
// - Then start using GET/POST requests to add data
// - Use redis-cli command to the GET "order:XXXX" and SMEMBERS orders

func main() {
	app := application.New()

	// NOTE: Create/derive a root Context.
	// Learn more about Context and how it can signal a graceful shutdown
	// whenever a SIGINT (i.e. Ctrl-C) is triggered (os.Interrupt)
	// REF: https://youtu.be/PWukxD1DC0I?t=472
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// NOTE: Could also use defer cancel() just under the initialization
	// of ctx, cancel. This defer will run at this end of the current
	// function it resides in; meaning I could call cancel() at end of main().
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Failed to start app:", err)
	}

}

// 'r' is a pointer of type http.Request (the inbound HTTP request from client)
// func basicHandler(w http.ResponseWriter, r *http.Request) {
// 	// NOTE: We could impl our own basic router like below, but this
// 	// is complicated handling path params, etc.
// 	// if r.Method == http.MethodGet {
// 	// 	// Handle GET
// 	// 	if r.URL.Path == "/foo" {
// 	// 		// Handle GET /foo
// 	// 	}
// 	// }
// 	//
// 	// if r.Method == http.MethodPost {
// 	// 	// Handle POST
// 	// }
//
// 	w.Write([]byte("Hello, world!"))
// }
