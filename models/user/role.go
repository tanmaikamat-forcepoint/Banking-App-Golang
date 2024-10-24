package user

import (
	"github.com/jinzhu/gorm"
)

type Role struct {
	gorm.Model        //RoleID
	RoleName   string `gorm:"not null" json:"roleName"` /// 'SUPER_ADMIN', 'BANK_USER', 'CLIENT_USER'
}
