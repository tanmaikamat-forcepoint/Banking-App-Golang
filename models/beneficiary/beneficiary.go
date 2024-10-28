package beneficiary

import (
	"bankManagement/models/client"
	"os/user"

	"github.com/jinzhu/gorm"
)

type Beneficiary struct {
	gorm.Model                   // Includes BeneficiaryID (ID), CreatedAt, UpdatedAt, DeletedAt
	BeneficiaryName       string `gorm:"not null" json:"beneficiary_name" validate:"required"`
	ClientID              uint   `gorm:"not null" json:"client_id"` // Foreign key - Client
	IsActive              bool   `gorm:"default:true" json:"is_active"`
	Status                string `gorm:"default:'Pending';not null" json:"status"`
	ApprovedByUserID      uint   `gorm:"column:approved_by_user_id" json:"approved_by_user_id"`       // Foreign key - Users
	CreatedByUserID       uint   `gorm:"column:created_by_user_id" json:"created_by_user_id"`         // Foreign key - Users
	BeneficiaryReceiverID uint   `gorm:"not null" json:"beneficiary_receiver_id" validate:"required"` // Foreign key - Client
	//  relationship - for foreign key
	ApprovedByUser      user.User     `gorm:"foreignkey:ApprovedByUserID;association_foreignkey:ID" json:"-"`
	CreatedByUser       user.User     `gorm:"foreignkey:CreatedByUserID;association_foreignkey:ID" json:"-"`
	Client              client.Client `gorm:"foreignkey:ClientID;association_foreignkey:ID" json:"-"`
	BeneficiaryReceiver client.Client `gorm:"foreignkey:BeneficiaryReceiverID;association_foreignkey:ID" json:"-"`
}
