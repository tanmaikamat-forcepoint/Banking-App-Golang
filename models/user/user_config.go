package user

import (
	"github.com/jinzhu/gorm"
)

type UserConfig struct {
	DB *gorm.DB
}

func (config *UserConfig) TableMigration() {
	config.DB.AutoMigrate(&User{}, &UserLoginInfo{})
	config.DB.Model(&User{}).AddForeignKey("role_id", "roles(id)", "CASCADE", "CASCADE")
	// config.DB.Model(&UserLoginInfo{}).AddForeignKey("user_id", "users(id)", "SET NULL", "CASCADE")
}
