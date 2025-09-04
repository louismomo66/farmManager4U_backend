package data

import "gorm.io/gorm"

type Models struct {
	User UserInterface
	Farm FarmInterface
}

func New(gormDB *gorm.DB) Models {
	return Models{
		User: NewUserRepo(gormDB),
		Farm: NewFarmRepo(gormDB),
	}
}
