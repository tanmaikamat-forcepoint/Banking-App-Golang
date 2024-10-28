package user

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique_index;not null" json:"username" validate:"required"` // unique username
	Password string `gorm:"not null" json:"password" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `gorm:"unique_index;not null" json:"email" validate:"required"` // Unique email
	IsActive bool   `gorm:"default:true" json:"is_active"`
	RoleID   uint   `gorm:"not null" json:"role_id"`
	Role     Role   `gorm:"foreignkey:RoleID;association_foreignkey:ID" json:"role"`
}

type UserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

type UserLoginParamDTO struct {
	Username string ` json:"username"  validate:"required"` // unique username
	Password string `json:"password"  validate:"required"`

}
