package user

import (
	"github.com/jinzhu/gorm"
)

type Role struct {
	gorm.Model
	RoleName string `gorm:"not null" json:"role_name"` //  'SUPER_ADMIN', 'BANK_USER', 'CLIENT_USER'
}
