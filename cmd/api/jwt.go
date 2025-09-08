package main

import (
	"errors"
	"farm4u/data"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT Claims structure
type Claims struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a JWT token for the user
func (app *Config) GenerateJWT(user *data.User) (string, error) {
	// Get JWT secret from environment variable, fallback to default
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key" // Change this in production!
	}

	// Get expiration time from environment variable, fallback to 24 hours
	expirationHours := 24
	if envExp := os.Getenv("JWT_EXPIRATION_HOURS"); envExp != "" {
		if hours, err := strconv.Atoi(envExp); err == nil {
			expirationHours = hours
		}
	}

	// Create claims
	claims := Claims{
		UserID: int(user.ID),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "farm4u",
			Subject:   strconv.Itoa(int(user.ID)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func (app *Config) ValidateJWT(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// JWT Middleware for protecting routes
func (app *Config) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.errorJSON(w, errors.New("authorization header required"), http.StatusUnauthorized)
			return
		}

		// Extract token (format: "Bearer <token>")
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			app.errorJSON(w, errors.New("invalid authorization header format"), http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := app.ValidateJWT(tokenString)
		if err != nil {
			app.errorJSON(w, errors.New("invalid or expired token"), http.StatusUnauthorized)
			return
		}

		// Add claims to request context for use in handlers
		r = r.WithContext(r.Context())
		r.Header.Set("X-User-ID", strconv.Itoa(claims.UserID))
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Role", claims.Role)

		next.ServeHTTP(w, r)
	}
}
