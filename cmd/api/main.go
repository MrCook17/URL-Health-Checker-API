package main

import (
	"log"
	"net/http"
	"time"

	"healthcheck-api/internal/httpapi"
	"healthcheck-api/internal/store"
)

func main() {
	// Create the in-memory job store.
	st := store.NewMemoryStore()

	// Create the HTTP handler layer with access to the store.
	h := httpapi.NewHandler(st)

	// Set up the standard library router and register API routes.
	mux := http.NewServeMux()
	h.Register(mux)

	// Configure the HTTP server.
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("server listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}