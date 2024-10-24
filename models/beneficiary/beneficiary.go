package models

import (
	"bankManagement/models/client"
	"os/user"

	"github.com/jinzhu/gorm"
)

type Beneficiary struct {
	gorm.Model                          //BeneficiaryID
	BeneficiaryName       string        `gorm:"type:varchar(255);not null"`
	ClientID              uint          `gorm:"not null"`
	Client                client.Client `gorm:"foreignkey:ClientID"`
	IsActive              bool          `gorm:"default:true"`
	Status                string        `gorm:"default:'Pending';not null"`
	ApprovedByUserID      uint          `gorm:"column:approved_by_user_id"`
	CreatedByUserID       uint          `gorm:"not null;column:created_by_user_id"`
	BeneficiaryReceiverID uint          `gorm:"not null"`
	BeneficiaryReceiver   client.Client `gorm:"foreignkey:BeneficiaryReceiverID"`
	User                  user.User     `gorm:"foreignKey:UserID;references:ID" json:"user"`
}
