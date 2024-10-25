package transaction

import (
	"bankManagement/models/client"

	"github.com/jinzhu/gorm"
)

type Transaction struct {
	gorm.Model                      //ID - TransactionID , CreatedAt
	ClientID          uint          `gorm:"not null" json:"client_id"`
	Client            client.Client `gorm:"foreignkey:ClientID;association_foreignkey:ID" json:"client"`
	PaymentType       string        `gorm:"not null" json:"payment_type"`     // Deposit, Withdraw, Transfer
	TransactionType   string        `gorm:"not null" json:"transaction_type"` // Credit, Debit
	TransactionAmount float64       `gorm:"not null" json:"transaction_amount"`
	TransactionStatus string        `gorm:"default:'Pending';not null" json:"transaction_status"`
}
