package client

import "os/user"

type ClientUser struct {
	UserID   uint `gorm:"primaryKey;autoIncrement:false" json:"userID"`
	ClientID uint `gorm:"primaryKey;autoIncrement:false" json:"clientID"`

	User   user.User `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Client Client    `gorm:"foreignKey:ClientID;references:ID" json:"client"`
}
