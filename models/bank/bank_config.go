package bank

import (
	"github.com/jinzhu/gorm"
)

type BankConfig struct {
	DB *gorm.DB
}

// TableMigration performs the migration for the Bank table
func (config *BankConfig) TableMigration() {
	// AutoMigrate will create the Bank table if it doesn't exist
	config.DB.AutoMigrate(&Bank{})
}
