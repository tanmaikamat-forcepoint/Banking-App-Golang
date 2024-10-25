package bank

import (
	"github.com/jinzhu/gorm"
)

type BankUserConfig struct {
	DB *gorm.DB
}

func (config *BankUserConfig) TableMigration() {
	config.DB.AutoMigrate(&BankUser{})

	config.DB.Model(&BankUser{}).AddUniqueIndex("idx_user_bank", "user_id", "bank_id") // uniques comp. primary key

	config.DB.Model(&BankUser{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	config.DB.Model(&BankUser{}).AddForeignKey("bank_id", "banks(id)", "CASCADE", "CASCADE")
}
