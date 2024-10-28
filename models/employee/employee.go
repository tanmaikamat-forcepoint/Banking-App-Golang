package employee

import (
	"bankManagement/models/client"

	"github.com/jinzhu/gorm"
)

type Employee struct {
	gorm.Model                        //  EmpID
	ClientID            uint          `gorm:"not null" json:"client_id"` // Foreign key to Client
	Client              client.Client `gorm:"foreignkey:ClientID;association_foreignkey:ID" json:"-"`
	EmployeeName        string        `gorm:"not null; default:'Test'" json:"name" validate:"required"`
	SalaryAmount        float64       `gorm:"not null" json:"salary_amount" validate:"required"`
	AccountNo           string        `gorm:"unique_index;not null" json:"account_no" validate:"required"`
	TotalSalaryReceived float64       `gorm:"default:0" json:"total_salary_received"`
}
