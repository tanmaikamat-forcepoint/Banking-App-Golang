package user

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model        //UserID
	Username   string `gorm:"uniqueIndex;not null" json:"username"`
	Password   string `gorm:"not null" json:"password"`
	Name       string `json:"name"`
	Email      string `gorm:"uniqueIndex;not null" json:"email"`
	IsActive   bool   `gorm:"default:true" json:"isActive"`
	RoleID     uint   `gorm:"not null" json:"roleID"`
	Role       Role   `gorm:"foreignKey:RoleID;references:ID" json:"role"`
}
