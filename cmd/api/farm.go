package main

import (
	"errors"
	"farm4u/data"
	"net/http"
)

// FarmRequest represents the farm creation/update request body
type FarmRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Size        float64 `json:"size"`
	FarmType    string  `json:"farmType"`
	Status      string  `json:"status"`
}

// FarmResponse represents the farm response
type FarmResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Farm    *data.Farm   `json:"farm,omitempty"`
	Farms   []*data.Farm `json:"farms,omitempty"`
}

// CreateFarmHandler handles farm creation
func (app *Config) CreateFarmHandler(w http.ResponseWriter, r *http.Request) {
	var req FarmRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Location == "" {
		app.errorJSON(w, errors.New("name and location are required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Validate farm type
	if req.FarmType == "" {
		req.FarmType = "Mixed" // Default farm type
	}

	// Validate status
	if req.Status == "" {
		req.Status = "Active" // Default status
	}

	// Validate size
	if req.Size <= 0 {
		app.errorJSON(w, errors.New("farm size must be greater than 0"), http.StatusBadRequest)
		return
	}

	// Get user from database using email from JWT claims
	user, err := app.Models.User.GetByEmail(userEmail)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil {
		app.errorJSON(w, errors.New("user not found"), http.StatusNotFound)
		return
	}

	// Create new farm
	farm := &data.Farm{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Size:        req.Size,
		FarmType:    req.FarmType,
		Status:      req.Status,
		UserID:      user.UserID, // Use the actual UserID from the user record
	}

	// Insert farm
	if err := app.Models.Farm.Insert(farm); err != nil {
		app.ErrorLog.Printf("Error creating farm: %v", err)
		app.errorJSON(w, errors.New("failed to create farm"), http.StatusInternalServerError)
		return
	}

	response := FarmResponse{
		Success: true,
		Message: "Farm created successfully",
		Farm:    farm,
	}

	app.writeJSON(w, http.StatusCreated, response)
}

// GetFarmHandler handles retrieving a single farm by ID
func (app *Config) GetFarmHandler(w http.ResponseWriter, r *http.Request) {
	// Get farm ID from URL parameters
	farmID := r.URL.Query().Get("id")
	if farmID == "" {
		app.errorJSON(w, errors.New("farm ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get farm by ID
	farm, err := app.Models.Farm.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil {
		app.errorJSON(w, errors.New("farm not found"), http.StatusNotFound)
		return
	}

	// Verify that the farm belongs to the authenticated user
	user, err := app.Models.User.GetByEmail(userEmail)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: farm does not belong to user"), http.StatusForbidden)
		return
	}

	response := FarmResponse{
		Success: true,
		Message: "Farm retrieved successfully",
		Farm:    farm,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// GetFarmsHandler handles retrieving all farms for a user
func (app *Config) GetFarmsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT claims (set by JWT middleware)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get user email from JWT claims to get the actual UserID
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get user from database to get the actual UserID
	user, err := app.Models.User.GetByEmail(userEmail)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil {
		app.errorJSON(w, errors.New("user not found"), http.StatusNotFound)
		return
	}

	// Get farms by user ID
	farms, err := app.Models.Farm.GetByUserID(user.UserID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farms: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	response := FarmResponse{
		Success: true,
		Message: "Farms retrieved successfully",
		Farms:   farms,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// UpdateFarmHandler handles farm updates
func (app *Config) UpdateFarmHandler(w http.ResponseWriter, r *http.Request) {
	var req FarmRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Get farm ID from URL parameters
	farmID := r.URL.Query().Get("id")
	if farmID == "" {
		app.errorJSON(w, errors.New("farm ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get existing farm
	existingFarm, err := app.Models.Farm.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if existingFarm == nil {
		app.errorJSON(w, errors.New("farm not found"), http.StatusNotFound)
		return
	}

	// Verify that the farm belongs to the authenticated user
	user, err := app.Models.User.GetByEmail(userEmail)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil || existingFarm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: farm does not belong to user"), http.StatusForbidden)
		return
	}

	// Update farm fields if provided
	if req.Name != "" {
		existingFarm.Name = req.Name
	}
	if req.Description != "" {
		existingFarm.Description = req.Description
	}
	if req.Location != "" {
		existingFarm.Location = req.Location
	}
	if req.Size > 0 {
		existingFarm.Size = req.Size
	}
	if req.FarmType != "" {
		existingFarm.FarmType = req.FarmType
	}
	if req.Status != "" {
		existingFarm.Status = req.Status
	}

	// Update farm
	if err := app.Models.Farm.Update(existingFarm); err != nil {
		app.ErrorLog.Printf("Error updating farm: %v", err)
		app.errorJSON(w, errors.New("failed to update farm"), http.StatusInternalServerError)
		return
	}

	response := FarmResponse{
		Success: true,
		Message: "Farm updated successfully",
		Farm:    existingFarm,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// DeleteFarmHandler handles farm deletion
func (app *Config) DeleteFarmHandler(w http.ResponseWriter, r *http.Request) {
	// Get farm ID from URL parameters
	farmID := r.URL.Query().Get("id")
	if farmID == "" {
		app.errorJSON(w, errors.New("farm ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get farm to verify it exists
	farm, err := app.Models.Farm.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil {
		app.errorJSON(w, errors.New("farm not found"), http.StatusNotFound)
		return
	}

	// Verify that the farm belongs to the authenticated user
	user, err := app.Models.User.GetByEmail(userEmail)
	if err != nil {
		app.ErrorLog.Printf("Error getting user by email: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if user == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: farm does not belong to user"), http.StatusForbidden)
		return
	}

	// Delete farm (soft delete)
	if err := app.Models.Farm.DeleteByID(int(farm.ID)); err != nil {
		app.ErrorLog.Printf("Error deleting farm: %v", err)
		app.errorJSON(w, errors.New("failed to delete farm"), http.StatusInternalServerError)
		return
	}

	response := FarmResponse{
		Success: true,
		Message: "Farm deleted successfully",
	}

	app.writeJSON(w, http.StatusOK, response)
}
