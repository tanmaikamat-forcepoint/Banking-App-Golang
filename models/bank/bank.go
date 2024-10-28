package bank

import (
	"github.com/jinzhu/gorm"
)

type Bank struct {
	gorm.Model              //BankID
	BankName         string `gorm:"unique_index;not null" json:"bank_name"`
	BankAbbreviation string `gorm:"not null" json:"bank_abbreviation"`
	IsActive         bool   `gorm:"default:true" json:"is_active"`
}

type BankAndUserDTO struct {
	BankName         string `json:"bank_name"`
	BankAbbreviation string `json:"bank_abbreviation"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	Email            string `json:"email"`
}
