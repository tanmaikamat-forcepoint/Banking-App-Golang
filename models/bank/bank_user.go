package bank

import (
	"os/user"
)

type BankUser struct {
	UserID uint `gorm:"primary_key;auto_increment:false" json:"user_id"`
	BankID uint `gorm:"primary_key;auto_increment:false" json:"bank_id"`

	Bank Bank      `gorm:"foreignkey:BankID;association_foreignkey:ID" json:"bank"`
	User user.User `gorm:"foreignkey:UserID;association_foreignkey:ID" json:"user"`
}
