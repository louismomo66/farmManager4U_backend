package main

import (
	"errors"
	"farm4u/data"
	"net/http"
	"time"
)

// EmployeeRequest represents the employee creation/update request body
type EmployeeRequest struct {
	UserID      *string    `json:"userId,omitempty"` // Optional link to User account
	FirstName   string     `json:"firstName"`
	LastName    string     `json:"lastName"`
	Position    string     `json:"position"`
	Salary      float64    `json:"salary"`
	HireDate    *time.Time `json:"hireDate"`
	ContactInfo string     `json:"contactInfo"`
	Status      string     `json:"status"`
}

// EmployeeResponse represents the employee response
type EmployeeResponse struct {
	Success   bool             `json:"success"`
	Message   string           `json:"message"`
	Employee  *data.Employee   `json:"employee,omitempty"`
	Employees []*data.Employee `json:"employees,omitempty"`
}

// CreateEmployeeHandler handles employee creation
func (app *Config) CreateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var req EmployeeRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Position == "" {
		app.errorJSON(w, errors.New("firstName, lastName, and position are required"), http.StatusBadRequest)
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

	// If UserID is provided, verify the user exists
	if req.UserID != nil && *req.UserID != "" {
		linkedUser, err := app.Models.User.GetByEmail(*req.UserID) // Assuming UserID is email for now
		if err != nil {
			app.ErrorLog.Printf("Error getting linked user: %v", err)
			app.errorJSON(w, errors.New("linked user not found"), http.StatusBadRequest)
			return
		}
		if linkedUser == nil {
			app.errorJSON(w, errors.New("linked user not found"), http.StatusBadRequest)
			return
		}
		req.UserID = &linkedUser.UserID // Use the actual UserID
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = "Active"
	}

	// Create new employee
	employee := &data.Employee{
		UserID:      req.UserID,
		FarmID:      farmID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Position:    req.Position,
		Salary:      req.Salary,
		HireDate:    req.HireDate,
		ContactInfo: req.ContactInfo,
		Status:      req.Status,
	}

	// Insert employee
	if err := app.Models.Employee.Insert(employee); err != nil {
		app.ErrorLog.Printf("Error creating employee: %v", err)
		app.errorJSON(w, errors.New("failed to create employee"), http.StatusInternalServerError)
		return
	}

	response := EmployeeResponse{
		Success:  true,
		Message:  "Employee created successfully",
		Employee: employee,
	}

	app.writeJSON(w, http.StatusCreated, response)
}

// GetEmployeeHandler handles retrieving a single employee by ID
func (app *Config) GetEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	// Get employee ID from URL parameters
	employeeID := r.URL.Query().Get("id")
	if employeeID == "" {
		app.errorJSON(w, errors.New("employee ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get employee by ID
	employee, err := app.Models.Employee.GetByEmployeeID(employeeID)
	if err != nil {
		app.ErrorLog.Printf("Error getting employee: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if employee == nil {
		app.errorJSON(w, errors.New("employee not found"), http.StatusNotFound)
		return
	}

	// Verify that the employee belongs to a farm owned by the authenticated user
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
	farm, err := app.Models.Farm.GetByFarmID(employee.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: employee does not belong to user's farm"), http.StatusForbidden)
		return
	}

	response := EmployeeResponse{
		Success:  true,
		Message:  "Employee retrieved successfully",
		Employee: employee,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// GetEmployeesHandler handles retrieving all employees for a farm
func (app *Config) GetEmployeesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get employees by farm ID
	employees, err := app.Models.Employee.GetByFarmID(farmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting employees: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	response := EmployeeResponse{
		Success:   true,
		Message:   "Employees retrieved successfully",
		Employees: employees,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// UpdateEmployeeHandler handles employee updates
func (app *Config) UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var req EmployeeRequest

	if err := app.ReadJSON(w, r, &req); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Get employee ID from URL parameters
	employeeID := r.URL.Query().Get("id")
	if employeeID == "" {
		app.errorJSON(w, errors.New("employee ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get existing employee
	existingEmployee, err := app.Models.Employee.GetByEmployeeID(employeeID)
	if err != nil {
		app.ErrorLog.Printf("Error getting employee: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if existingEmployee == nil {
		app.errorJSON(w, errors.New("employee not found"), http.StatusNotFound)
		return
	}

	// Verify that the employee belongs to a farm owned by the authenticated user
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
	farm, err := app.Models.Farm.GetByFarmID(existingEmployee.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: employee does not belong to user's farm"), http.StatusForbidden)
		return
	}

	// If UserID is provided, verify the user exists
	if req.UserID != nil && *req.UserID != "" {
		linkedUser, err := app.Models.User.GetByEmail(*req.UserID) // Assuming UserID is email for now
		if err != nil {
			app.ErrorLog.Printf("Error getting linked user: %v", err)
			app.errorJSON(w, errors.New("linked user not found"), http.StatusBadRequest)
			return
		}
		if linkedUser == nil {
			app.errorJSON(w, errors.New("linked user not found"), http.StatusBadRequest)
			return
		}
		req.UserID = &linkedUser.UserID // Use the actual UserID
	}

	// Update employee fields if provided
	if req.FirstName != "" {
		existingEmployee.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existingEmployee.LastName = req.LastName
	}
	if req.Position != "" {
		existingEmployee.Position = req.Position
	}
	if req.Salary > 0 {
		existingEmployee.Salary = req.Salary
	}
	if req.HireDate != nil {
		existingEmployee.HireDate = req.HireDate
	}
	if req.ContactInfo != "" {
		existingEmployee.ContactInfo = req.ContactInfo
	}
	if req.Status != "" {
		existingEmployee.Status = req.Status
	}
	if req.UserID != nil {
		existingEmployee.UserID = req.UserID
	}

	// Update employee
	if err := app.Models.Employee.Update(existingEmployee); err != nil {
		app.ErrorLog.Printf("Error updating employee: %v", err)
		app.errorJSON(w, errors.New("failed to update employee"), http.StatusInternalServerError)
		return
	}

	response := EmployeeResponse{
		Success:  true,
		Message:  "Employee updated successfully",
		Employee: existingEmployee,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// DeleteEmployeeHandler handles employee deletion
func (app *Config) DeleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	// Get employee ID from URL parameters
	employeeID := r.URL.Query().Get("id")
	if employeeID == "" {
		app.errorJSON(w, errors.New("employee ID is required"), http.StatusBadRequest)
		return
	}

	// Get user email from JWT claims (set by JWT middleware)
	userEmail := r.Header.Get("X-User-Email")
	if userEmail == "" {
		app.errorJSON(w, errors.New("user not authenticated"), http.StatusUnauthorized)
		return
	}

	// Get employee to verify it exists
	employee, err := app.Models.Employee.GetByEmployeeID(employeeID)
	if err != nil {
		app.ErrorLog.Printf("Error getting employee: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if employee == nil {
		app.errorJSON(w, errors.New("employee not found"), http.StatusNotFound)
		return
	}

	// Verify that the employee belongs to a farm owned by the authenticated user
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
	farm, err := app.Models.Farm.GetByFarmID(employee.FarmID)
	if err != nil {
		app.ErrorLog.Printf("Error getting farm: %v", err)
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	if farm == nil || farm.UserID != user.UserID {
		app.errorJSON(w, errors.New("access denied: employee does not belong to user's farm"), http.StatusForbidden)
		return
	}

	// Delete employee (soft delete)
	if err := app.Models.Employee.DeleteByID(int(employee.ID)); err != nil {
		app.ErrorLog.Printf("Error deleting employee: %v", err)
		app.errorJSON(w, errors.New("failed to delete employee"), http.StatusInternalServerError)
		return
	}

	response := EmployeeResponse{
		Success: true,
		Message: "Employee deleted successfully",
	}

	app.writeJSON(w, http.StatusOK, response)
}
