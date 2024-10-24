package bank

import "github.com/jinzhu/gorm"

type BankUser struct {
	gorm.Model
	BankId uint
	UserId uint
}
