package document

import (
	"bankManagement/models/client"
	"bankManagement/models/user"

	"github.com/jinzhu/gorm"
)

type Document struct {
	gorm.Model
	FileName         string        `gorm:"not null"`
	FileType         string        `gorm:"not null"`
	FileURL          string        `gorm:"not null"` // store file path
	UploadedByUserId uint          `gorm:"not null"`
	ClientId         uint          `gorm:"not null"`
	Client           client.Client `gorm:"foreignkey:ClientId"`
	UploadedByUser   user.User     `gorm:"foreignkey:UploadedByUserId"`
}
