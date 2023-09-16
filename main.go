package main

import (
	"context"
	"fmt"

	"github.com/gaylonalfano/go-turso-htmx-orders/application"
)

func main() {
	app := application.New()

	err := app.Start(context.TODO())
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
