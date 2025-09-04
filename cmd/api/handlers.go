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

// SignupRequest represents the signup request body
type SignupRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	PhoneNumber string `json:"phoneNumber"`
	Address     string `json:"address"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	User    *data.User `json:"user,omitempty"`
	Token   string     `json:"token,omitempty"`
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

// SignupHandler handles user registration
func (app *Config) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		app.errorJSON(w, errors.New("missing required fields"), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, err := app.Models.User.GetByEmail(req.Email)
	if err != nil {
		app.ErrorLog.Printf("Error checking existing user: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		app.errorJSON(w, errors.New("user with this email already exists"), http.StatusConflict)
		return
	}

	// Create new user
	user := &data.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		TempPassword: req.Password,
		Role:         req.Role,
		PhoneNumber:  req.PhoneNumber,
		Address:      req.Address,
		Active:       true,
	}

	// Insert user (password will be hashed automatically)
	if err := app.Models.User.Insert(user); err != nil {
		app.ErrorLog.Printf("Error creating user: %v", err)
		app.errorJSON(w, errors.New("failed to create user"), http.StatusInternalServerError)
		return
	}

	// Clear sensitive data before sending response
	user.Password = ""
	user.TempPassword = ""

	response := AuthResponse{
		Success: true,
		Message: "User created successfully",
		User:    user,
	}

	app.writeJSON(w, http.StatusCreated, response)
}

// LoginHandler handles user authentication
func (app *Config) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		app.errorJSON(w, errors.New("email and password are required"), http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := app.Models.User.GetByEmail(req.Email)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil {
		app.errorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !user.Active {
		app.errorJSON(w, errors.New("account is deactivated"), http.StatusUnauthorized)
		return
	}

	// Verify password
	matches, err := app.Models.User.PasswordMatches(user, req.Password)
	if err != nil {
		app.ErrorLog.Printf("Error checking password: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if !matches {
		app.errorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := app.GenerateJWT(user)
	if err != nil {
		app.ErrorLog.Printf("Error generating JWT token: %v", err)
		app.errorJSON(w, errors.New("failed to generate authentication token"), http.StatusInternalServerError)
		return
	}

	// Clear sensitive data before sending response
	user.Password = ""
	user.TempPassword = ""

	response := AuthResponse{
		Success: true,
		Message: "Login successful",
		User:    user,
		Token:   token,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// ForgotPasswordHandler handles password reset requests
func (app *Config) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		app.errorJSON(w, errors.New("email is required"), http.StatusBadRequest)
		return
	}

	// Check if user exists
	user, err := app.Models.User.GetByEmail(req.Email)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Don't reveal if user exists or not for security
		response := AuthResponse{
			Success: true,
			Message: "If the email exists, a password reset code has been sent",
		}
		app.writeJSON(w, http.StatusOK, response)
		return
	}

	// Generate OTP
	otp, err := app.Models.User.GenerateAndSaveOTP(req.Email)
	if err != nil {
		app.ErrorLog.Printf("Error generating OTP: %v", err)
		app.errorJSON(w, errors.New("failed to generate reset code"), http.StatusInternalServerError)
		return
	}

	// TODO: Send OTP via email/SMS
	app.InfoLog.Printf("OTP for %s: %s", req.Email, otp)

	response := AuthResponse{
		Success: true,
		Message: "Password reset code has been sent to your email",
	}

	app.writeJSON(w, http.StatusOK, response)
}

// ResetPasswordHandler handles password reset with OTP
func (app *Config) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"newPassword"`
	}

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		app.errorJSON(w, errors.New("email, OTP, and new password are required"), http.StatusBadRequest)
		return
	}

	// Reset password with OTP
	if err := app.Models.User.ResetPasswordWithOTP(req.Email, req.OTP, req.NewPassword); err != nil {
		app.ErrorLog.Printf("Error resetting password: %v", err)
		app.errorJSON(w, errors.New("invalid or expired reset code"), http.StatusBadRequest)
		return
	}

	response := AuthResponse{
		Success: true,
		Message: "Password reset successfully",
	}

	app.writeJSON(w, http.StatusOK, response)
}

// RefreshTokenHandler generates a new JWT token for authenticated users
func (app *Config) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get current user from token (assumes JWT middleware was used)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Convert userID to int
	id, err := strconv.Atoi(userID)
	if err != nil {
		app.errorJSON(w, errors.New("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := app.Models.User.GetOne(id)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by ID: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil || !user.Active {
		app.errorJSON(w, errors.New("user not found or inactive"), http.StatusUnauthorized)
		return
	}

	// Generate new JWT token
	token, err := app.GenerateJWT(user)
	if err != nil {
		app.ErrorLog.Printf("Error generating JWT token: %v", err)
		app.errorJSON(w, errors.New("failed to generate authentication token"), http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Token:   token,
	}

	app.writeJSON(w, http.StatusOK, response)
}
