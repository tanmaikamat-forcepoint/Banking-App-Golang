package bank

import (
	"github.com/jinzhu/gorm"
)

type Bank struct {
	gorm.Model  //BankID
	BankName   string `gorm:"not null" json:"bankName"`
	BankAbbrev string `gorm:"not null" json:"bankAbbrev"`
	IsActive   bool   `gorm:"default:true" json:"isActive"`
}
