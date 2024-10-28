package user

import (
	"github.com/jinzhu/gorm"
)

type RoleConfig struct {
	DB *gorm.DB
}

func (config *RoleConfig) TableMigration() {
	config.DB.AutoMigrate(&Role{})

}
