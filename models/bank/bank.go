package bank

import (
	"github.com/jinzhu/gorm"
)

type Bank struct {
	gorm.Model              //BankID
	BankName         string `gorm:"not null" json:"bankName"`
	BankAbbreviation string `gorm:"not null" json:"bank_abbreviation"`
	IsActive         bool   `gorm:"default:true" json:"is_active"`
}
