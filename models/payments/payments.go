package payments

import "github.com/jinzhu/gorm"

type Payment struct {
	gorm.Model
	SenderClientID      uint    `gorm:"not null"`
	ReceiverClientID    uint    `gorm:"not null"`
	AuthorizedBankId    uint    `gorm:"not null"`
	CreditTransactionID uint    `gorm:"not null"`
	DebitTransactionID  uint    `gorm:"not null"`
	Amount              float64 `gorm:"not null"`
	Status              string  `gorm:"default:'Pending'"` //// 'Pending', 'Approved', 'Rejected'
	CreatedByUserId     uint    `gorm:"not null"`
	ApprovedByUserId    uint    `gorm:"not null"`
}

type PaymentRequest struct {
	gorm.Model
	SenderClientID   uint    `gorm:"not null"`
	ReceiverClientID uint    `gorm:"not null"`
	AuthorizerBankId uint    `gorm:"not null"`
	Amount           float64 `gorm:"not null"`
	CreatedByUserId  uint    `gorm:"not null"`
}
type PaymentRequestDTO struct {
	Amount        float64 `json:"amount"`
	BeneficiaryId uint    `json:"beneficiary_id"`
}

type PaymentResponseDTO struct {
	Amount        float64 `json:"amount"`
	ClientId      uint    `json:"client_id"`
	PaymentId     uint    `json:"payment_id"`
	PaymentStatus uint    `json:"payment_status"`
	Timestamp     string  `json:"created_at"`
}
