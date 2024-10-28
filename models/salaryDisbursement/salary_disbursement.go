package salaryDisbursement

import (
	"bankManagement/models/client"
	"bankManagement/models/employee"
	"bankManagement/models/user"

	"github.com/jinzhu/gorm"
)

type SalaryDisbursement struct {
	gorm.Model
	ClientID        uint              `gorm:"not null"`
	Client          client.Client     `gorm:"foreignkey:ClientID"`
	EmpID           uint              `gorm:"not null"`
	Employee        employee.Employee `gorm:"foreignkey:EmpID"`
	TransactionID   uint              `gorm:"not null"`
	SalaryAmount    float64           `gorm:"not null"`
	Status          string            `gorm:"default:'Pending'"`
	CreatedByUserId uint              `gorm:"not null"`
	CreatedByUser   user.User         `gorm:"foreignkey:CreatedByUserId"`
}
