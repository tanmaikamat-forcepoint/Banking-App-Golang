package transaction

import (
	"github.com/jinzhu/gorm"
)

type TransactionConfig struct {
	DB *gorm.DB
}

func (config *TransactionConfig) TableMigration() {
	config.DB.AutoMigrate(&Transaction{})

	config.DB.Model(&Transaction{}).AddForeignKey("client_id", "clients(id)", "CASCADE", "CASCADE")
}
