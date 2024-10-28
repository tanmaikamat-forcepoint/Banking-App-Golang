package encrypt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var signingKey = "This is Amazing Key"

var AdminUserRoleID = 1
var BankUserRoleID = 2
var ClientUserRoleID = 3

type Claims struct {
	UserId       uint `json:"userId"`
	RoleId       uint `json:"roleId"`
	BankId       uint `json:"bankId,omitempty"`
	ClientId     uint `json:"clientId,omitempty"`
	IsSuperAdmin bool `json:"is_super_admin,omitempty"`
	jwt.StandardClaims
}

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func CheckHashWithPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func GetJwtFromData(userId uint, RoleId uint, BankId uint, ClientId uint, isSuperAdmin bool) (string, error) {
	claims := &Claims{UserId: userId, RoleId: RoleId, BankId: BankId, ClientId: ClientId, IsSuperAdmin: isSuperAdmin}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	finalToken, err := token.SignedString([]byte(signingKey))
	return finalToken, err
}

func ValidateJwtToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("Invalid token")
	}
	return claims, nil
}
