package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()
	//specify who is allowed to connect
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(middleware.Heartbeat("/ping"))

	// Health check endpoint
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Authentication routes
	mux.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", app.SignupHandler)
		r.Post("/login", app.LoginHandler)
		r.Post("/forgot-password", app.ForgotPasswordHandler)
		r.Post("/reset-password", app.ResetPasswordHandler)
		r.Post("/refresh-token", app.JWTMiddleware(app.RefreshTokenHandler))
	})

	// Farm routes (protected with JWT middleware)
	mux.Route("/api/farms", func(r chi.Router) {
		r.Post("/", app.JWTMiddleware(app.CreateFarmHandler))
		r.Get("/", app.JWTMiddleware(app.GetFarmsHandler))
		r.Get("/{id}", app.JWTMiddleware(app.GetFarmHandler))
		r.Put("/{id}", app.JWTMiddleware(app.UpdateFarmHandler))
		r.Delete("/{id}", app.JWTMiddleware(app.DeleteFarmHandler))
	})

	return mux
}
