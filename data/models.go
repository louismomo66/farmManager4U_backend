package data

import "gorm.io/gorm"

type Models struct {
	User      UserInterface
	Farm      FarmInterface
	Crop      CropInterface
	Livestock LivestockInterface
	Employee  EmployeeInterface
}

func New(gormDB *gorm.DB) Models {
	return Models{
		User:      NewUserRepo(gormDB),
		Farm:      NewFarmRepo(gormDB),
		Crop:      NewCropRepo(gormDB),
		Livestock: NewLivestockRepo(gormDB),
		Employee:  NewEmployeeRepo(gormDB),
	}
}
