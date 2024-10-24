package transaction

import (
	"bankManagement/models/client"
	"time"

	"github.com/jinzhu/gorm"
)

type Transaction struct {
	gorm.Model                      //TransactionID
	ClientID          uint          `gorm:"not null" json:"clientID"`
	Client            client.Client `gorm:"foreignKey:ClientID;references:ID" json:"client"`
	PaymentType       string        `gorm:"not null" json:"paymentType"`     // Deposit, Withdraw, Transfer
	TransactionType   string        `gorm:"not null" json:"transactionType"` // credit, debit
	TransactionAmount float64       `gorm:"not null" json:"transactionAmount"`
	Timestamp         time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"timestamp"`
	TransactionStatus string        `gorm:"default:'Pending';not null" json:"transactionStatus"`
}
