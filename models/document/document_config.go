package document

import "github.com/jinzhu/gorm"

type DocumentConfig struct {
	DB *gorm.DB
}

func (tconf *DocumentConfig) TableMigration() {
	tconf.DB.AutoMigrate(&Document{})
}
