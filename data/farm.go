package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Farm represents the farms table in the database.
type Farm struct {
	ID          uint           `gorm:"primaryKey" json:"-"`
	FarmID      string         `gorm:"primaryKey;size:36;default:gen_random_uuid()" json:"farmId"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Location    string         `gorm:"not null" json:"location"`
	Size        float64        `gorm:"not null" json:"size"`                    // Size in acres/hectares
	FarmType    string         `gorm:"not null" json:"farmType"`                // e.g., "Crop", "Livestock", "Mixed"
	Status      string         `gorm:"not null;default:'Active'" json:"status"` // Active, Inactive, Suspended
	UserID      string         `gorm:"not null;size:36" json:"userId"`          // Foreign key to User
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
}

// FarmRepo implements FarmInterface using GORM.
type FarmRepo struct {
	DB *gorm.DB
}

// NewFarmRepo creates a new instance of FarmRepo.
func NewFarmRepo(db *gorm.DB) FarmInterface {
	return &FarmRepo{DB: db}
}

// GetAll retrieves all farms from the database
func (f *FarmRepo) GetAll() ([]*Farm, error) {
	var farms []*Farm
	result := f.DB.Find(&farms)
	return farms, result.Error
}

// GetByID retrieves a farm by its ID
func (f *FarmRepo) GetByID(id int) (*Farm, error) {
	var farm Farm
	result := f.DB.Where("id = ?", id).First(&farm)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &farm, result.Error
}

// GetByFarmID retrieves a farm by its FarmID (UUID)
func (f *FarmRepo) GetByFarmID(farmID string) (*Farm, error) {
	var farm Farm
	result := f.DB.Where("farm_id = ?", farmID).First(&farm)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &farm, result.Error
}

// GetByUserID retrieves all farms belonging to a specific user
func (f *FarmRepo) GetByUserID(userID string) ([]*Farm, error) {
	var farms []*Farm
	result := f.DB.Where("user_id = ?", userID).Find(&farms)
	return farms, result.Error
}

// Insert creates a new farm in the database
func (f *FarmRepo) Insert(farm *Farm) error {
	return f.DB.Create(farm).Error
}

// Update updates an existing farm in the database
func (f *FarmRepo) Update(farm *Farm) error {
	return f.DB.Save(farm).Error
}

// DeleteByID soft deletes a farm by its ID
func (f *FarmRepo) DeleteByID(id int) error {
	return f.DB.Delete(&Farm{}, id).Error
}
