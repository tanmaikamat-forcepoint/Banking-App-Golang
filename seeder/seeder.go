package seeder

import (
	"bankManagement/models/user"
	"fmt"

	"github.com/jinzhu/gorm"
)

func SeedRoles(db *gorm.DB) {
	fmt.Println("Roles has been initialized....")
	roles := []user.Role{
		{RoleName: "SUPER_ADMIN"},
		{RoleName: "BANK_USER"},
		{RoleName: "CLIENT_USER"},
	}

	for _, role := range roles {
		var existingRole user.Role
		if err := db.Where("role_name = ?", role.RoleName).First(&existingRole).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				if err := db.Create(&role).Error; err == nil {
					fmt.Printf("Inserted role: %s\n", role.RoleName)
				} else {
					fmt.Printf("Error inserting role: %s\n", err)
				}
			}
		}
	}
}
