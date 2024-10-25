package user

import (
	"github.com/jinzhu/gorm"
)

type UserConfig struct {
	DB *gorm.DB
}

func (config *UserConfig) TableMigration() {
	config.DB.AutoMigrate(&User{})
}
