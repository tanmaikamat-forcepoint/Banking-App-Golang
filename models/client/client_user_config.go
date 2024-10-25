package client

import (
	"github.com/jinzhu/gorm"
)

type ClientUserConfig struct {
	DB *gorm.DB
}

func (config *ClientUserConfig) TableMigration() {
	config.DB.AutoMigrate(&ClientUser{})

	config.DB.Model(&ClientUser{}).AddUniqueIndex("idx_user_client", "user_id", "client_id") // unique prim. comp. key

	config.DB.Model(&ClientUser{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	config.DB.Model(&ClientUser{}).AddForeignKey("client_id", "clients(id)", "CASCADE", "CASCADE")
}
