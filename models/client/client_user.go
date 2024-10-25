package client

import "os/user"

type ClientUser struct {
	UserID   uint `gorm:"primary_key;auto_increment:false" json:"user_id"`
	ClientID uint `gorm:"primary_key;auto_increment:false" json:"client_id"`

	User   user.User `gorm:"foreignkey:UserID;association_foreignkey:ID" json:"user"`
	Client Client    `gorm:"foreignkey:ClientID;association_foreignkey:ID" json:"client"`
}
