package main

import (
	"errors"
	"farm4u/data"
	"net/http"
	"time"
)

// CropRequest represents the crop creation/update request body
type CropRequest struct {
	Name         string     `json:"name"`
	PlantingDate *time.Time `json:"plantingDate"`
	HarvestDate  *time.Time `json:"harvestDate"`
	Quantity     float64    `json:"quantity"`
	Status       string     `json:"status"`
	Notes        string     `json:"notes"`
}

// CropResponse represents the crop response
type CropResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Crop    *data.Crop   `json:"crop,omitempty"`
	Crops   []*data.Crop `json:"crops,omitempty"`
}

// CreateCropHandler handles crop creation
func (app *Config) CreateCropHandler(w http.ResponseWriter, r *http.Request) {
	var req CropRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Quantity <= 0 {
		app.errorJSON(w, errors.New("name and quantity are required"), http.StatusBadRequest)
		return
	}

	// Get farm ID from URL parameters
	farmID := r.URL.Query().Get("farmId")
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

	// Verify that the farm belongs to the authenticated user
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

	// Verify farm exists and belongs to user
	farm, err := app.Models.Farm.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("farm not found or access denied"), http.StatusForbidden)
		return
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = "Growing"
	}

	// Create new crop
	crop := &data.Crop{
		FarmID:       farmID,
		Name:         req.Name,
		PlantingDate: req.PlantingDate,
		HarvestDate:  req.HarvestDate,
		Quantity:     req.Quantity,
		Status:       req.Status,
		Notes:        req.Notes,
	}

	// Insert crop
	if err := app.Models.Crop.Insert(crop); err != nil {
		app.ErrorLog.Printf("Error creating crop: %v", err)
		app.errorJSON(w, errors.New("failed to create crop"), http.StatusInternalServerError)
		return
	}

	response := CropResponse{
		Success: true,
		Message: "Crop created successfully",
		Crop:    crop,
	}

	app.writeJSON(w, http.StatusCreated, response)
}

// GetCropHandler handles retrieving a single crop by ID
func (app *Config) GetCropHandler(w http.ResponseWriter, r *http.Request) {
	// Get crop ID from URL parameters
	cropID := r.URL.Query().Get("id")
	if cropID == "" {
		app.errorJSON(w, errors.New("crop ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get crop by ID
	crop, err := app.Models.Crop.GetByCropID(cropID)
	if err != nil {
		app.ErrorLog.Printf("Error getting crop: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if crop == nil {
		app.errorJSON(w, errors.New("crop not found"), http.StatusNotFound)
		return
	}

	// Verify that the crop belongs to a farm owned by the authenticated user
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

	// Get the farm to verify ownership
	farm, err := app.Models.Farm.GetByFarmID(crop.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: crop does not belong to user's farm"), http.StatusForbidden)
		return
	}

	response := CropResponse{
		Success: true,
		Message: "Crop retrieved successfully",
		Crop:    crop,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// GetCropsHandler handles retrieving all crops for a farm
func (app *Config) GetCropsHandler(w http.ResponseWriter, r *http.Request) {
	// Get farm ID from URL parameters
	farmID := r.URL.Query().Get("farmId")
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

	// Verify that the farm belongs to the authenticated user
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

	// Verify farm exists and belongs to user
	farm, err := app.Models.Farm.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("farm not found or access denied"), http.StatusForbidden)
		return
	}

	// Get crops by farm ID
	crops, err := app.Models.Crop.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting crops: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	response := CropResponse{
		Success: true,
		Message: "Crops retrieved successfully",
		Crops:   crops,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// UpdateCropHandler handles crop updates
func (app *Config) UpdateCropHandler(w http.ResponseWriter, r *http.Request) {
	var req CropRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Get crop ID from URL parameters
	cropID := r.URL.Query().Get("id")
	if cropID == "" {
		app.errorJSON(w, errors.New("crop ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get existing crop
	existingCrop, err := app.Models.Crop.GetByCropID(cropID)
	if err != nil {
		app.ErrorLog.Printf("Error getting crop: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if existingCrop == nil {
		app.errorJSON(w, errors.New("crop not found"), http.StatusNotFound)
		return
	}

	// Verify that the crop belongs to a farm owned by the authenticated user
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

	// Get the farm to verify ownership
	farm, err := app.Models.Farm.GetByFarmID(existingCrop.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: crop does not belong to user's farm"), http.StatusForbidden)
		return
	}

	// Update crop fields if provided
	if req.Name != "" {
		existingCrop.Name = req.Name
	}
	if req.PlantingDate != nil {
		existingCrop.PlantingDate = req.PlantingDate
	}
	if req.HarvestDate != nil {
		existingCrop.HarvestDate = req.HarvestDate
	}
	if req.Quantity > 0 {
		existingCrop.Quantity = req.Quantity
	}
	if req.Status != "" {
		existingCrop.Status = req.Status
	}
	if req.Notes != "" {
		existingCrop.Notes = req.Notes
	}

	// Update crop
	if err := app.Models.Crop.Update(existingCrop); err != nil {
		app.ErrorLog.Printf("Error updating crop: %v", err)
		app.errorJSON(w, errors.New("failed to update crop"), http.StatusInternalServerError)
		return
	}

	response := CropResponse{
		Success: true,
		Message: "Crop updated successfully",
		Crop:    existingCrop,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// DeleteCropHandler handles crop deletion
func (app *Config) DeleteCropHandler(w http.ResponseWriter, r *http.Request) {
	// Get crop ID from URL parameters
	cropID := r.URL.Query().Get("id")
	if cropID == "" {
		app.errorJSON(w, errors.New("crop ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get crop to verify it exists
	crop, err := app.Models.Crop.GetByCropID(cropID)
	if err != nil {
		app.ErrorLog.Printf("Error getting crop: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if crop == nil {
		app.errorJSON(w, errors.New("crop not found"), http.StatusNotFound)
		return
	}

	// Verify that the crop belongs to a farm owned by the authenticated user
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

	// Get the farm to verify ownership
	farm, err := app.Models.Farm.GetByFarmID(crop.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: crop does not belong to user's farm"), http.StatusForbidden)
		return
	}

	// Delete crop (soft delete)
	if err := app.Models.Crop.DeleteByID(int(crop.ID)); err != nil {
		app.ErrorLog.Printf("Error deleting crop: %v", err)
		app.errorJSON(w, errors.New("failed to delete crop"), http.StatusInternalServerError)
		return
	}

	response := CropResponse{
		Success: true,
		Message: "Crop deleted successfully",
	}

	app.writeJSON(w, http.StatusOK, response)
}
