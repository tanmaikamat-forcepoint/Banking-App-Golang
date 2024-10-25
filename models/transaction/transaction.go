package transaction

import "github.com/jinzhu/gorm"

type Transaction struct {
	gorm.Model
	ClientID          uint    `gorm:"not null"`
	PaymentType       string  `gorm:"not null"` // Deposit,Withdraw, 'Transfer'
	TransactionType   string  `gorm:"not null"` //credit , debit
	TransactionAmount float64 `gorm:"not null"`
	TransactionStatus string  `gorm:"not null"`
}
