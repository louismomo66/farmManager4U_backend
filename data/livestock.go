package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Livestock represents the livestock table in the database.
type Livestock struct {
	ID              uint           `gorm:"primaryKey" json:"-"`
	LivestockID     string         `gorm:"primaryKey;size:36;default:gen_random_uuid()" json:"livestockId"`
	FarmID          string         `gorm:"not null;size:36" json:"farmId"` // Foreign key to Farm
	Type            string         `gorm:"not null" json:"type"`           // Cattle, Poultry, Sheep, Goat, etc.
	Count           int            `gorm:"not null" json:"count"`          // Number of animals
	AcquisitionDate *time.Time     `json:"acquisitionDate"`
	HealthStatus    string         `gorm:"not null;default:'Healthy'" json:"healthStatus"` // Healthy, Sick, Under Treatment, Deceased
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Farm *Farm `gorm:"foreignKey:FarmID;references:FarmID" json:"farm,omitempty"`
}

// LivestockInterface defines the contract for livestock operations
type LivestockInterface interface {
	GetAll() ([]*Livestock, error)
	GetByID(id int) (*Livestock, error)
	GetByLivestockID(livestockID string) (*Livestock, error)
	GetByFarmID(farmID string) ([]*Livestock, error)
	Insert(livestock *Livestock) error
	Update(livestock *Livestock) error
	DeleteByID(id int) error
	GetByType(livestockType string) ([]*Livestock, error)
	GetByHealthStatus(healthStatus string) ([]*Livestock, error)
}

// LivestockRepo implements LivestockInterface using GORM.
type LivestockRepo struct {
	DB *gorm.DB
}

// NewLivestockRepo creates a new instance of LivestockRepo.
func NewLivestockRepo(db *gorm.DB) LivestockInterface {
	return &LivestockRepo{DB: db}
}

// GetAll retrieves all livestock from the database
func (l *LivestockRepo) GetAll() ([]*Livestock, error) {
	var livestock []*Livestock
	result := l.DB.Find(&livestock)
	return livestock, result.Error
}

// GetByID retrieves a livestock by its ID
func (l *LivestockRepo) GetByID(id int) (*Livestock, error) {
	var livestock Livestock
	result := l.DB.Where("id = ?", id).First(&livestock)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &livestock, result.Error
}

// GetByLivestockID retrieves a livestock by its LivestockID (UUID)
func (l *LivestockRepo) GetByLivestockID(livestockID string) (*Livestock, error) {
	var livestock Livestock
	result := l.DB.Where("livestock_id = ?", livestockID).First(&livestock)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &livestock, result.Error
}

// GetByFarmID retrieves all livestock belonging to a specific farm
func (l *LivestockRepo) GetByFarmID(farmID string) ([]*Livestock, error) {
	var livestock []*Livestock
	result := l.DB.Where("farm_id = ?", farmID).Find(&livestock)
	return livestock, result.Error
}

// GetByType retrieves all livestock of a specific type
func (l *LivestockRepo) GetByType(livestockType string) ([]*Livestock, error) {
	var livestock []*Livestock
	result := l.DB.Where("type = ?", livestockType).Find(&livestock)
	return livestock, result.Error
}

// GetByHealthStatus retrieves all livestock with a specific health status
func (l *LivestockRepo) GetByHealthStatus(healthStatus string) ([]*Livestock, error) {
	var livestock []*Livestock
	result := l.DB.Where("health_status = ?", healthStatus).Find(&livestock)
	return livestock, result.Error
}

// Insert creates a new livestock in the database
func (l *LivestockRepo) Insert(livestock *Livestock) error {
	return l.DB.Create(livestock).Error
}

// Update updates an existing livestock in the database
func (l *LivestockRepo) Update(livestock *Livestock) error {
	return l.DB.Save(livestock).Error
}

// DeleteByID soft deletes a livestock by its ID
func (l *LivestockRepo) DeleteByID(id int) error {
	return l.DB.Delete(&Livestock{}, id).Error
}
