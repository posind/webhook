package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gocroot/route"
)

// Define the logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register your handler
	mux.HandleFunc("/", route.URL)

	// Wrap the mux with the logging middleware
	loggedMux := loggingMiddleware(mux)

	// Start the server
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", loggedMux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
