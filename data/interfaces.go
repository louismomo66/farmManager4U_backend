package data

type UserInterface interface {
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	GetOne(id int) (*User, error)
	Update(user *User) error
	Insert(user *User) error
	ResetPassword(password string, user User) error
	DeleteByID(id int) error
	PasswordMatches(user *User, plainText string) (bool, error)
	GenerateAndSaveOTP(email string) (string, error)
	VerifyOTP(email, otp string) (bool, error)
	ResetPasswordWithOTP(email, otp, newPassword string) error
}
