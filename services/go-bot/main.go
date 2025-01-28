package main

import (
	"log"
	"net/http"
	"os"

	"github.com/david-botos/BearHug/services/go-bot/handlers"
	"github.com/joho/godotenv"
)

func main() {
	// Set the network port
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	port, exists := os.LookupEnv("PORT")

	if !exists {
		port = "3001"
	}

	// Use default mux for routing
	http.HandleFunc("/dial-out", handlers.DialOutHandler)
	http.HandleFunc("/webhook", handlers.WebhookHandler)

	// Start the server
	log.Printf("The server is running at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
