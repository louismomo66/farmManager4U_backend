package main

import (
	"errors"
	"farm4u/data"
	"net/http"
	"time"
)

// LivestockRequest represents the livestock creation/update request body
type LivestockRequest struct {
	Type            string     `json:"type"`
	Count           int        `json:"count"`
	AcquisitionDate *time.Time `json:"acquisitionDate"`
	HealthStatus    string     `json:"healthStatus"`
	Notes           string     `json:"notes"`
}

// LivestockResponse represents the livestock response
type LivestockResponse struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Livestock  *data.Livestock   `json:"livestock,omitempty"`
	Livestocks []*data.Livestock `json:"livestocks,omitempty"`
}

// CreateLivestockHandler handles livestock creation
func (app *Config) CreateLivestockHandler(w http.ResponseWriter, r *http.Request) {
	var req LivestockRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Type == "" || req.Count <= 0 {
		app.errorJSON(w, errors.New("type and count are required"), http.StatusBadRequest)
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

	// Set default health status if not provided
	if req.HealthStatus == "" {
		req.HealthStatus = "Healthy"
	}

	// Create new livestock
	livestock := &data.Livestock{
		FarmID:          farmID,
		Type:            req.Type,
		Count:           req.Count,
		AcquisitionDate: req.AcquisitionDate,
		HealthStatus:    req.HealthStatus,
		Notes:           req.Notes,
	}

	// Insert livestock
	if err := app.Models.Livestock.Insert(livestock); err != nil {
		app.ErrorLog.Printf("Error creating livestock: %v", err)
		app.errorJSON(w, errors.New("failed to create livestock"), http.StatusInternalServerError)
		return
	}

	response := LivestockResponse{
		Success:   true,
		Message:   "Livestock created successfully",
		Livestock: livestock,
	}

	app.writeJSON(w, http.StatusCreated, response)
}

// GetLivestockHandler handles retrieving a single livestock by ID
func (app *Config) GetLivestockHandler(w http.ResponseWriter, r *http.Request) {
	// Get livestock ID from URL parameters
	livestockID := r.URL.Query().Get("id")
	if livestockID == "" {
		app.errorJSON(w, errors.New("livestock ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get livestock by ID
	livestock, err := app.Models.Livestock.GetByLivestockID(livestockID)
	if err != nil {
		app.ErrorLog.Printf("Error getting livestock: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if livestock == nil {
		app.errorJSON(w, errors.New("livestock not found"), http.StatusNotFound)
		return
	}

	// Verify that the livestock belongs to a farm owned by the authenticated user
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
	farm, err := app.Models.Farm.GetByFarmID(livestock.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: livestock does not belong to user's farm"), http.StatusForbidden)
		return
	}

	response := LivestockResponse{
		Success:   true,
		Message:   "Livestock retrieved successfully",
		Livestock: livestock,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// GetLivestocksHandler handles retrieving all livestock for a farm
func (app *Config) GetLivestocksHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get livestock by farm ID
	livestocks, err := app.Models.Livestock.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting livestock: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	response := LivestockResponse{
		Success:    true,
		Message:    "Livestock retrieved successfully",
		Livestocks: livestocks,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// UpdateLivestockHandler handles livestock updates
func (app *Config) UpdateLivestockHandler(w http.ResponseWriter, r *http.Request) {
	var req LivestockRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Get livestock ID from URL parameters
	livestockID := r.URL.Query().Get("id")
	if livestockID == "" {
		app.errorJSON(w, errors.New("livestock ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get existing livestock
	existingLivestock, err := app.Models.Livestock.GetByLivestockID(livestockID)
	if err != nil {
		app.ErrorLog.Printf("Error getting livestock: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if existingLivestock == nil {
		app.errorJSON(w, errors.New("livestock not found"), http.StatusNotFound)
		return
	}

	// Verify that the livestock belongs to a farm owned by the authenticated user
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
	farm, err := app.Models.Farm.GetByFarmID(existingLivestock.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: livestock does not belong to user's farm"), http.StatusForbidden)
		return
	}

	// Update livestock fields if provided
	if req.Type != "" {
		existingLivestock.Type = req.Type
	}
	if req.Count > 0 {
		existingLivestock.Count = req.Count
	}
	if req.AcquisitionDate != nil {
		existingLivestock.AcquisitionDate = req.AcquisitionDate
	}
	if req.HealthStatus != "" {
		existingLivestock.HealthStatus = req.HealthStatus
	}
	if req.Notes != "" {
		existingLivestock.Notes = req.Notes
	}

	// Update livestock
	if err := app.Models.Livestock.Update(existingLivestock); err != nil {
		app.ErrorLog.Printf("Error updating livestock: %v", err)
		app.errorJSON(w, errors.New("failed to update livestock"), http.StatusInternalServerError)
		return
	}

	response := LivestockResponse{
		Success:   true,
		Message:   "Livestock updated successfully",
		Livestock: existingLivestock,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// DeleteLivestockHandler handles livestock deletion
func (app *Config) DeleteLivestockHandler(w http.ResponseWriter, r *http.Request) {
	// Get livestock ID from URL parameters
	livestockID := r.URL.Query().Get("id")
	if livestockID == "" {
		app.errorJSON(w, errors.New("livestock ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get livestock to verify it exists
	livestock, err := app.Models.Livestock.GetByLivestockID(livestockID)
	if err != nil {
		app.ErrorLog.Printf("Error getting livestock: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if livestock == nil {
		app.errorJSON(w, errors.New("livestock not found"), http.StatusNotFound)
		return
	}

	// Verify that the livestock belongs to a farm owned by the authenticated user
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
	farm, err := app.Models.Farm.GetByFarmID(livestock.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: livestock does not belong to user's farm"), http.StatusForbidden)
		return
	}

	// Delete livestock (soft delete)
	if err := app.Models.Livestock.DeleteByID(int(livestock.ID)); err != nil {
		app.ErrorLog.Printf("Error deleting livestock: %v", err)
		app.errorJSON(w, errors.New("failed to delete livestock"), http.StatusInternalServerError)
		return
	}

	response := LivestockResponse{
		Success: true,
		Message: "Livestock deleted successfully",
	}

	app.writeJSON(w, http.StatusOK, response)
}
