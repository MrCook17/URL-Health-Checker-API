package main

import (
	"log"
	"net/http"
	"time"

	"healthcheck-api/internal/httpapi"
	"healthcheck-api/internal/store"
)

func main() {
	st := store.NewMemoryStore()
	h := httpapi.NewHandler(st)

	mux := http.NewServeMux()
	h.Register(mux)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("server listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}