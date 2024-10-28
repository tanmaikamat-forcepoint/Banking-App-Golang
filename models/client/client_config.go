package client

import (
	"github.com/jinzhu/gorm"
)

type ClientConfig struct {
	DB *gorm.DB
}

func (config *ClientConfig) TableMigration() {
	config.DB.AutoMigrate(&Client{})

	config.DB.Model(&Client{}).AddForeignKey("bank_id", "banks(id)", "CASCADE", "CASCADE")
}
