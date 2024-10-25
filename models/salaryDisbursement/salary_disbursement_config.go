package salaryDisbursement

import "github.com/jinzhu/gorm"

type SalaryDisbursementConfig struct {
	DB *gorm.DB
}

func (tconf *SalaryDisbursementConfig) TableMigration() {
	tconf.DB.AutoMigrate(&SalaryDisbursement{})
}
