package employee

import (
	"bankManagement/models/client"

	"github.com/jinzhu/gorm"
)

type Employee struct {
	gorm.Model                        //  EmpID
	ClientID            uint          `gorm:"not null" json:"client_id"`
	Client              client.Client `gorm:"foreignKey:ClientID" json:"client"`
	SalaryAmount        float64       `gorm:"not null" json:"salary_amount"`
	AccountNo           string        `gorm:"uniqueIndex;not null" json:"account_no"`
	TotalSalaryReceived float64       `gorm:"default:0" json:"total_salary_received"`
}
