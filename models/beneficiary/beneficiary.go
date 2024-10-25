package beneficiary

import (
	"bankManagement/models/client"
	"os/user"

	"github.com/jinzhu/gorm"
)

type Beneficiary struct {
	gorm.Model                   // Includes BeneficiaryID (ID), CreatedAt, UpdatedAt, DeletedAt
	BeneficiaryName       string `gorm:"not null" json:"beneficiary_name"`
	ClientID              uint   `gorm:"not null" json:"client_id"` // Foreign key - Client
	IsActive              bool   `gorm:"default:true" json:"is_active"`
	Status                string `gorm:"default:'Pending';not null" json:"status"`
	ApprovedByUserID      *uint  `gorm:"column:approved_by_user_id" json:"approved_by_user_id"` // Foreign key - Users
	CreatedByUserID       *uint  `gorm:"column:created_by_user_id" json:"created_by_user_id"`   // Foreign key - Users
	BeneficiaryReceiverID uint   `gorm:"not null" json:"beneficiary_receiver_id"`               // Foreign key - Client
	//  relationship - for foreign key
	ApprovedByUser      user.User     `gorm:"foreignkey:ApprovedByUserID;association_foreignkey:ID" json:"approved_by_user"`
	CreatedByUser       user.User     `gorm:"foreignkey:CreatedByUserID;association_foreignkey:ID" json:"created_by_user"`
	Client              client.Client `gorm:"foreignkey:ClientID;association_foreignkey:ID" json:"client"`
	BeneficiaryReceiver client.Client `gorm:"foreignkey:BeneficiaryReceiverID;association_foreignkey:ID" json:"beneficiary_receiver"`
}
