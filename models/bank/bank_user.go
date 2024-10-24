package bank

import "os/user"

type BankUser struct {
	BankID uint `gorm:"not null;index" json:"bank_id"`
	UserID uint `gorm:"not null;index" json:"user_id"`

	Bank Bank      `gorm:"foreignKey:BankID" json:"bank"`
	User user.User `gorm:"foreignKey:UserID" json:"user"`
}
