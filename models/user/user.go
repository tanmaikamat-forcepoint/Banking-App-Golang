package user

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Username string
}
type UserConfig struct {
	DB *gorm.DB
}

func (conf *UserConfig) TableMigration() {
	conf.DB.AutoMigrate(&User{})
}
