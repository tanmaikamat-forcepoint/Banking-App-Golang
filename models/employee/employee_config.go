package employee

import (
	"github.com/jinzhu/gorm"
)

type EmployeeConfig struct {
	DB *gorm.DB
}

func (config *EmployeeConfig) TableMigration() {
	config.DB.AutoMigrate(&Employee{})

	config.DB.Model(&Employee{}).AddForeignKey("client_id", "clients(id)", "CASCADE", "CASCADE")
}
