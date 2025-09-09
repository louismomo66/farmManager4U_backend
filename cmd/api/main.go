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

	app := Config{
		InfoLog:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLog: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	db := app.initDB()
	if db == nil {
		app.ErrorLog.Fatal("Failed to initialize database")
	}

	// Initialize models
	models := data.New(db)

	app.DB = db
	app.Models = models

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app.routes(),
	}

	app.InfoLog.Printf("Starting Farm Manager 4U API server on port %d", port)
	app.InfoLog.Printf("Database connected successfully")
	app.InfoLog.Printf("API endpoints available at http://localhost:%d", port)
	app.InfoLog.Printf("Health check: http://localhost:%d/health", port)

	if err := srv.ListenAndServe(); err != nil {
		app.ErrorLog.Fatal("Failed to start server:", err)
	}
}
