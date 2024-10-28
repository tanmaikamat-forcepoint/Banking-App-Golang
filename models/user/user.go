package user

import (
	"time"

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

type UserLoginInfo struct {
	gorm.Model
	UserId    string    `gorm:"not null" json:"user_id"`
	UserName  string    `gorm:"not null" json:"username"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	RoleID    uint      `gorm:"not null" json:"role_id"`
	LoginTime time.Time `gorm:"not null" json:"login_time"`
	Role      Role      `gorm:"foreignkey:RoleID;association_foreignkey:ID" json:"role"`
}
type UserLoginParamDTO struct {
	Username string ` json:"username"  validate:"required"` // unique username
	Password string `json:"password"  validate:"required"`
}

type UserPermissionDTO struct {
	BankId       uint `json:"bank_id"  ` // unique username
	ClientId     uint `json:"client_id" `
	IsSuperAdmin bool `json:"is_super_admin"`
}
