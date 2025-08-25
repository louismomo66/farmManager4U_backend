package data

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents the users table in the database.
type User struct {
	ID           uint           `gorm:"primaryKey" json:"-"`
	UserID       string         `gorm:"primaryKey;size:36;default:gen_random_uuid()" json:"userId"`
	FirstName    string         `gorm:"not null" json:"firstName"`
	LastName     string         `gorm:"not null" json:"lastName"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	Password     string         `gorm:"not null" json:"-"`
	TempPassword string         `json:"password" gorm:"-"` // Temporary field for password unmarshaling
	Role         string         `gorm:"not null;default:'Farmer'" json:"role"`
	PhoneNumber  string         `json:"phoneNumber"`
	Address      string         `json:"address"`
	Active       bool           `gorm:"default:true" json:"active"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	// OTP fields
	OTPCode      string    `gorm:"type:varchar(6)" json:"-"`
	OTPExpiresAt time.Time `json:"-"`
}

// UserRepo implements UserInterface using GORM.
type UserRepo struct {
	DB *gorm.DB
}

// NewUserRepo creates a new instance of UserRepo.
func NewUserRepo(db *gorm.DB) UserInterface {
	return &UserRepo{DB: db}
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// GetAll retrieves all users from the database
func (u *UserRepo) GetAll() ([]*User, error) {
	var users []*User
	result := u.DB.Find(&users)
	return users, result.Error
}

// GetByEmail retrieves a user by their email address
func (u *UserRepo) GetByEmail(email string) (*User, error) {
	var user User
	result := u.DB.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, result.Error
}

// GetOne retrieves a user by their ID
func (u *UserRepo) GetOne(id int) (*User, error) {
	var user User
	result := u.DB.Where("id = ?", id).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, result.Error
}

// Insert creates a new user in the database after hashing the password
func (u *UserRepo) Insert(user *User) error {
	// Hash the password before saving
	hashedPassword, err := HashPassword(user.TempPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return u.DB.Create(user).Error
}

// Update updates an existing user in the database
func (u *UserRepo) Update(user *User) error {
	// If the password is being updated, hash it
	if user.TempPassword != "" {
		hashedPassword, err := HashPassword(user.TempPassword)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	} else {
		// Prevent overwriting the password with an empty string
		var existingUser User
		if err := u.DB.First(&existingUser, user.ID).Error; err != nil {
			return err
		}
		user.Password = existingUser.Password
	}

	return u.DB.Save(user).Error
}

// ResetPassword updates the password for a specific user
func (u *UserRepo) ResetPassword(password string, user User) error {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	return u.DB.Model(&user).Update("password", hashedPassword).Error
}

// DeleteByID soft deletes a user by their ID
func (u *UserRepo) DeleteByID(id int) error {
	return u.DB.Delete(&User{}, id).Error
}

// PasswordMatches checks if the provided plain text password matches the stored hashed password
func (u *UserRepo) PasswordMatches(user *User, plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainText))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GenerateAndSaveOTP generates a new OTP code for the user and saves it to the database
func (u *UserRepo) GenerateAndSaveOTP(email string) (string, error) {
	var user User
	result := u.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return "", result.Error
	}

	// Generate a random 6-digit OTP using crypto/rand for better security
	otpNum := 100000 + rand.New(rand.NewSource(time.Now().UnixNano())).Intn(900000)
	otp := strconv.Itoa(otpNum)

	// Set OTP and expiration (15 minutes from now)
	user.OTPCode = otp
	user.OTPExpiresAt = time.Now().Add(15 * time.Minute)

	// Save the user with the new OTP
	if err := u.DB.Save(&user).Error; err != nil {
		return "", err
	}

	return otp, nil
}

// VerifyOTP checks if the provided OTP is valid for the user
func (u *UserRepo) VerifyOTP(email, otp string) (bool, error) {
	var user User
	result := u.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return false, result.Error
	}

	// Check if OTP matches and has not expired
	if user.OTPCode != otp {
		return false, nil
	}

	if time.Now().After(user.OTPExpiresAt) {
		return false, errors.New("OTP has expired")
	}

	return true, nil
}

// ResetPasswordWithOTP resets a user's password after validating the OTP
func (u *UserRepo) ResetPasswordWithOTP(email, otp, newPassword string) error {
	// Verify OTP first
	valid, err := u.VerifyOTP(email, otp)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("invalid or expired OTP")
	}

	var user User
	if err := u.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update the password and clear the OTP
	user.Password = hashedPassword
	user.OTPCode = ""

	// Save the changes
	return u.DB.Save(&user).Error
}
