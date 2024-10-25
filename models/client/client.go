package client

import (
	"bankManagement/models/bank"

	"github.com/jinzhu/gorm"
)

type Client struct {
	gorm.Model                   // ClientID
	ClientName         string    `gorm:"not null" json:"client_name"`
	ClientEmail        string    `gorm:"unique_index;not null" json:"client_email"`
	Balance            float64   `gorm:"default:1000" json:"balance"`
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	VerificationStatus string    `gorm:"default:'Pending';not null" json:"verification_status"`
	BankID             uint      `gorm:"not null;index" json:"bank_id"`
	Bank               bank.Bank `gorm:"foreignkey:BankID;association_foreignkey:ID" json:"bank"`
}
