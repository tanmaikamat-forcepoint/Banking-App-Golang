package salaryDisbursement

import "github.com/jinzhu/gorm"

type SalaryDisbursement struct {
	gorm.Model
	ClientID         uint    `gorm:"not null"`
	EmpID            uint    `gorm:"not null"`
	TransactionID    uint    `gorm:"not null"`
	SalaryAmount     float64 `gorm:"not null"`
	Status           string  `gorm:"default:'Pending'"`
	CreatedByUserId  uint    `gorm:"not null"`
	ApprovedByUserId uint
}
