package client

import (
	"bankManagement/models/bank"

	"github.com/jinzhu/gorm"
)

type Client struct {
	gorm.Model                   // Clientgit ID
	ClientName         string    `gorm:"not null" json:"clientName"`
	ClientEmail        string    `gorm:"uniqueIndex;not null" json:"clientEmail"`
	Balance            float64   `gorm:"default:1000" json:"balance"`
	IsActive           bool      `gorm:"default:true" json:"isActive"`
	VerificationStatus string    `gorm:"default:'Pending';not null" json:"verificationStatus"`
	BankID             uint      `gorm:"not null;index" json:"bank_id"`
	Bank               bank.Bank `gorm:"foreignKey:BankID" json:"bank"`
}
