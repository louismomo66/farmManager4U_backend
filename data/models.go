package data

import "gorm.io/gorm"

type Models struct {
	User UserInterface
}

func New(gormDB *gorm.DB) Models {
	return Models{
		User: NewUserRepo(gormDB),
	}
}
