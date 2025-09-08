package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Employee represents the employees table in the database.
type Employee struct {
	ID          uint           `gorm:"primaryKey" json:"-"`
	EmployeeID  string         `gorm:"primaryKey;size:36;default:gen_random_uuid()" json:"employeeId"`
	UserID      *string        `gorm:"size:36" json:"userId,omitempty"` // Optional foreign key to User (nullable)
	FarmID      string         `gorm:"not null;size:36" json:"farmId"`  // Foreign key to Farm
	FirstName   string         `gorm:"not null" json:"firstName"`
	LastName    string         `gorm:"not null" json:"lastName"`
	Position    string         `gorm:"not null" json:"position"` // Job title or role
	Salary      float64        `json:"salary"`                   // Compensation details
	HireDate    *time.Time     `json:"hireDate"`
	ContactInfo string         `json:"contactInfo"`                             // Phone or email for contact
	Status      string         `gorm:"not null;default:'Active'" json:"status"` // Active, Inactive, Terminated
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
	Farm *Farm `gorm:"foreignKey:FarmID;references:FarmID" json:"farm,omitempty"`
}

// EmployeeInterface defines the contract for employee operations
type EmployeeInterface interface {
	GetAll() ([]*Employee, error)
	GetByID(id int) (*Employee, error)
	GetByEmployeeID(employeeID string) (*Employee, error)
	GetByFarmID(farmID string) ([]*Employee, error)
	GetByUserID(userID string) ([]*Employee, error)
	Insert(employee *Employee) error
	Update(employee *Employee) error
	DeleteByID(id int) error
	GetByPosition(position string) ([]*Employee, error)
	GetByStatus(status string) ([]*Employee, error)
}

// EmployeeRepo implements EmployeeInterface using GORM.
type EmployeeRepo struct {
	DB *gorm.DB
}

// NewEmployeeRepo creates a new instance of EmployeeRepo.
func NewEmployeeRepo(db *gorm.DB) EmployeeInterface {
	return &EmployeeRepo{DB: db}
}

// GetAll retrieves all employees from the database
func (e *EmployeeRepo) GetAll() ([]*Employee, error) {
	var employees []*Employee
	result := e.DB.Find(&employees)
	return employees, result.Error
}

// GetByID retrieves an employee by its ID
func (e *EmployeeRepo) GetByID(id int) (*Employee, error) {
	var employee Employee
	result := e.DB.Where("id = ?", id).First(&employee)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &employee, result.Error
}

// GetByEmployeeID retrieves an employee by its EmployeeID (UUID)
func (e *EmployeeRepo) GetByEmployeeID(employeeID string) (*Employee, error) {
	var employee Employee
	result := e.DB.Where("employee_id = ?", employeeID).First(&employee)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &employee, result.Error
}

// GetByFarmID retrieves all employees belonging to a specific farm
func (e *EmployeeRepo) GetByFarmID(farmID string) ([]*Employee, error) {
	var employees []*Employee
	result := e.DB.Where("farm_id = ?", farmID).Find(&employees)
	return employees, result.Error
}

// GetByUserID retrieves all employees linked to a specific user
func (e *EmployeeRepo) GetByUserID(userID string) ([]*Employee, error) {
	var employees []*Employee
	result := e.DB.Where("user_id = ?", userID).Find(&employees)
	return employees, result.Error
}

// GetByPosition retrieves all employees with a specific position
func (e *EmployeeRepo) GetByPosition(position string) ([]*Employee, error) {
	var employees []*Employee
	result := e.DB.Where("position = ?", position).Find(&employees)
	return employees, result.Error
}

// GetByStatus retrieves all employees with a specific status
func (e *EmployeeRepo) GetByStatus(status string) ([]*Employee, error) {
	var employees []*Employee
	result := e.DB.Where("status = ?", status).Find(&employees)
	return employees, result.Error
}

// Insert creates a new employee in the database
func (e *EmployeeRepo) Insert(employee *Employee) error {
	return e.DB.Create(employee).Error
}

// Update updates an existing employee in the database
func (e *EmployeeRepo) Update(employee *Employee) error {
	return e.DB.Save(employee).Error
}

// DeleteByID soft deletes an employee by its ID
func (e *EmployeeRepo) DeleteByID(id int) error {
	return e.DB.Delete(&Employee{}, id).Error
}
