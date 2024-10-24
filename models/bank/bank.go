package bank

import "github.com/jinzhu/gorm"

type Bank struct {
	gorm.Model
	BankName     string
	Abbreviation string
	isActive     bool
}
