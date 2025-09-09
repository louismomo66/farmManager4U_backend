package main

import (
	"farm4u/data"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Set default port
	port := 9005
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := fmt.Sscanf(envPort, "%d", &port); err != nil || p != 1 {
			log.Printf("Invalid PORT environment variable, using default port %d", port)
		}
	}

	app := Config{}

	db := app.initDB()
	if db == nil {
		log.Fatal("Failed to initialize database")
	}

	// Initialize models
	models := data.New(db)

	app.DB = db
	app.Models = models

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app.routes(),
	}

	log.Printf("Starting server on port %d", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
