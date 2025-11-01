package main

import (
	"log"
	"net/http"
	"os"

	"github.com/edpsouza/chatterbox/internal/config"
	"github.com/edpsouza/chatterbox/internal/handlers"
	"github.com/edpsouza/chatterbox/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize SQLite store
	storeInstance, err := store.NewStore(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer storeInstance.Close()
	// Set global store instance for WebSocket authentication
	handlers.SetStoreInstance(storeInstance)

	// Initialize WebSocket hub
	hub := handlers.NewHub()
	go hub.Run()

	// Set up HTTP routes
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeWS(hub, w, r)
	})
	http.HandleFunc("/register", handlers.RegisterHandler(storeInstance))
	http.HandleFunc("/login", handlers.LoginHandler(storeInstance))
	http.HandleFunc("/users/", handlers.PublicKeyHandler(storeInstance))

	log.Printf("Starting server on port %s...", cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
