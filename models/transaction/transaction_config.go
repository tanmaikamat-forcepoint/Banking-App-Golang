package transaction

import "github.com/jinzhu/gorm"

type TransactionConfig struct {
	DB *gorm.DB
}

func (tconf *TransactionConfig) TableMigration() {
	tconf.DB.AutoMigrate(&Transaction{})
}
