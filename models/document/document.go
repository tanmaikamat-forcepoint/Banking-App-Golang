package document

import (
	"bankManagement/models/client"
	"bankManagement/models/user"

	"github.com/jinzhu/gorm"
)

type Document struct {
	gorm.Model
	FileName         string        `gorm:"not null" json:"file_name"`
	FileType         string        `gorm:"not null" json:"file_type"`
	FileURL          string        `gorm:"not null" json:"file_url"` // store file path
	UploadedByUserId uint          `gorm:"not null" json:"uploaded_by_user_id"`
	ClientId         uint          `gorm:"not null" json:"client_id"`
	Client           client.Client `gorm:"foreignkey:ClientId" json:"client"`
	UploadedByUser   user.User     `gorm:"foreignkey:UploadedByUserId" json:"uploaded_by_user"`
}
