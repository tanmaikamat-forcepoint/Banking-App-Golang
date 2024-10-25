package beneficiary

import (
	"github.com/jinzhu/gorm"
)

type BeneficiaryConfig struct {
	DB *gorm.DB
}

func (config *BeneficiaryConfig) TableMigration() {
	config.DB.AutoMigrate(&Beneficiary{})

	config.DB.Model(&Beneficiary{}).AddForeignKey("client_id", "clients(id)", "CASCADE", "CASCADE")
	config.DB.Model(&Beneficiary{}).AddForeignKey("approved_by_user_id", "users(id)", "SET NULL", "CASCADE")
	config.DB.Model(&Beneficiary{}).AddForeignKey("created_by_user_id", "users(id)", "SET NULL", "CASCADE")
	config.DB.Model(&Beneficiary{}).AddForeignKey("beneficiary_receiver_id", "clients(id)", "CASCADE", "CASCADE")
}
