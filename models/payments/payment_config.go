package payments

import "github.com/jinzhu/gorm"

type PaymentConfig struct {
	DB *gorm.DB
}

func (tconf *PaymentConfig) TableMigration() {
	tconf.DB.AutoMigrate(&Payment{}, &PaymentRequest{})
}
