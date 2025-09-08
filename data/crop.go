package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Crop represents the crops table in the database.
type Crop struct {
	ID           uint           `gorm:"primaryKey" json:"-"`
	CropID       string         `gorm:"primaryKey;size:36;default:gen_random_uuid()" json:"cropId"`
	FarmID       string         `gorm:"not null;size:36" json:"farmId"` // Foreign key to Farm
	Name         string         `gorm:"not null" json:"name"`
	PlantingDate *time.Time     `json:"plantingDate"`
	HarvestDate  *time.Time     `json:"harvestDate"`
	Quantity     float64        `gorm:"not null" json:"quantity"`                 // Amount planted (kg or number of plants)
	Status       string         `gorm:"not null;default:'Growing'" json:"status"` // Growing, Harvested, Failed
	Notes        string         `json:"notes"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Farm *Farm `gorm:"foreignKey:FarmID;references:FarmID" json:"farm,omitempty"`
}

// CropInterface defines the contract for crop operations
type CropInterface interface {
	GetAll() ([]*Crop, error)
	GetByID(id int) (*Crop, error)
	GetByCropID(cropID string) (*Crop, error)
	GetByFarmID(farmID string) ([]*Crop, error)
	Insert(crop *Crop) error
	Update(crop *Crop) error
	DeleteByID(id int) error
	GetByStatus(status string) ([]*Crop, error)
}

// CropRepo implements CropInterface using GORM.
type CropRepo struct {
	DB *gorm.DB
}

// NewCropRepo creates a new instance of CropRepo.
func NewCropRepo(db *gorm.DB) CropInterface {
	return &CropRepo{DB: db}
}

// GetAll retrieves all crops from the database
func (c *CropRepo) GetAll() ([]*Crop, error) {
	var crops []*Crop
	result := c.DB.Find(&crops)
	return crops, result.Error
}

// GetByID retrieves a crop by its ID
func (c *CropRepo) GetByID(id int) (*Crop, error) {
	var crop Crop
	result := c.DB.Where("id = ?", id).First(&crop)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &crop, result.Error
}

// GetByCropID retrieves a crop by its CropID (UUID)
func (c *CropRepo) GetByCropID(cropID string) (*Crop, error) {
	var crop Crop
	result := c.DB.Where("crop_id = ?", cropID).First(&crop)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &crop, result.Error
}

// GetByFarmID retrieves all crops belonging to a specific farm
func (c *CropRepo) GetByFarmID(farmID string) ([]*Crop, error) {
	var crops []*Crop
	result := c.DB.Where("farm_id = ?", farmID).Find(&crops)
	return crops, result.Error
}

// GetByStatus retrieves all crops with a specific status
func (c *CropRepo) GetByStatus(status string) ([]*Crop, error) {
	var crops []*Crop
	result := c.DB.Where("status = ?", status).Find(&crops)
	return crops, result.Error
}

// Insert creates a new crop in the database
func (c *CropRepo) Insert(crop *Crop) error {
	return c.DB.Create(crop).Error
}

// Update updates an existing crop in the database
func (c *CropRepo) Update(crop *Crop) error {
	return c.DB.Save(crop).Error
}

// DeleteByID soft deletes a crop by its ID
func (c *CropRepo) DeleteByID(id int) error {
	return c.DB.Delete(&Crop{}, id).Error
}
