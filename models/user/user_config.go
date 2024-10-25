package user

import (
	"bankManagement/models/bank"

	"github.com/jinzhu/gorm"
)

type UserConfig struct {
	DB *gorm.DB
}

func (config *UserConfig) TableMigration() {
	config.DB.AutoMigrate(&User{})
	config.DB.Table("users").Related(&bank.Bank{})
}
