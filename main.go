package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	router := chi.NewRouter()

	// Storing 'server' as a pointer, which means we're storing the memory
	// address, NOT as a value!
	server := &http.Server{
		Addr:    ":3000",
		Handler: http.HandlerFunc(basicHandler),
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Failed to listen to server", err)
	}
}

// 'r' is a pointer of type http.Request (the inbound HTTP request from client)
func basicHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE: We could impl our own basic router like below, but this
	// is complicated handling path params, etc.
	// if r.Method == http.MethodGet {
	// 	// Handle GET
	// 	if r.URL.Path == "/foo" {
	// 		// Handle GET /foo
	// 	}
	// }
	//
	// if r.Method == http.MethodPost {
	// 	// Handle POST
	// }

	w.Write([]byte("Hello, world!"))
}
