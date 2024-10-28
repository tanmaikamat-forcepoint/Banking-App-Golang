package payments

import (
	"bankManagement/models/bank"
	"bankManagement/models/client"
	"bankManagement/models/transaction"
	"os/user"

	"github.com/jinzhu/gorm"
)

type Payment struct {
	gorm.Model
	SenderClientID      uint                    `gorm:"not null"`
	SenderClient        client.Client           `gorm:"foreignkey:SenderClientID`
	ReceiverClientID    uint                    `gorm:"not null"`
	ReceiverClient      client.Client           `gorm:"foreignkey:ReceiverClientID`
	AuthorizedBankID    uint                    `gorm:"not null"`
	AuthorizedBank      bank.Bank               `gorm:"foreignkey:AuthorizedBankID`
	CreditTransactionID uint                    `gorm:"not null"`
	CreditTransaction   transaction.Transaction `gorm:"foreignkey:CreditTransactionID`
	DebitTransactionID  uint                    `gorm:"not null"`
	DebitTransaction    transaction.Transaction `gorm:"foreignkey:DebitTransactionID`
	Amount              float64                 `gorm:"not null"`
	Status              string                  `gorm:"default:'Pending'"` //// 'Pending', 'Approved', 'Rejected'
	CreatedByUserId     uint                    `gorm:"not null"`
	ApprovedByUserId    uint                    `gorm:"not null"`
	CreatedByUser       user.User               `gorm:"foreignkey:CreatedByUserId`
	ApprovedByUser      user.User               `gorm:"foreignkey:ApprovedByUserId`
}

type PaymentRequest struct {
	gorm.Model
	SenderClientID   uint          `gorm:"not null" json:"received_client_id"`
	SenderClient     client.Client `gorm:"foreignkey:SenderClientID" json:"-"`
	ReceiverClientID uint          `gorm:"not null"  json:"received_client_id"`
	ReceiverClient   client.Client `gorm:"foreignkey:ReceiverClientID" json:"receiver"`
	AuthorizerBankId uint          `gorm:"not null"  json:"authorized_bank_id"`
	AuthorizedBank   bank.Bank     `gorm:"foreignkey:AuthorizedBankID" json:"-"`
	Amount           float64       `gorm:"not null" json:"amount"`
	Resolved         bool          `gorm:"default:false" json:"resolved"`
	CreatedByUserId  uint          `gorm:"not null" json:"create_by"`
	CreatedByUser    user.User     `gorm:"foreignkey:CreatedByUserId" json:"-"`
}
type PaymentRequestDTO struct {
	Amount        float64 `json:"amount" validate:"required"`
	BeneficiaryId uint    `json:"beneficiary_id"`
}

type PaymentResponseDTO struct {
	Amount        float64 `json:"amount"`
	ClientId      uint    `json:"client_id"`
	PaymentId     uint    `json:"payment_id"`
	PaymentStatus uint    `json:"payment_status"`
	Timestamp     string  `json:"created_at"`
}
