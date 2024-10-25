package document

import "github.com/jinzhu/gorm"

type Document struct {
	gorm.Model
	FileName         string `gorm:"not null"`
	FileType         string `gorm:"not null"`
	FileURL          string `gorm:"not null"` // store file path
	UploadedByUserId uint   `gorm:"not null"`
}
