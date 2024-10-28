package auth

import (
	"bankManagement/constants"
	"bankManagement/utils/encrypt"
	errorsUtils "bankManagement/utils/errors"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Authentication Middleware Called")
		token, err := getAuthTokenFromHeader(r)
		if err != nil {
			errorsUtils.SendInvalidAuthError(w)
			return
		}
		claims, err1 := encrypt.ValidateJwtToken(token)
		fmt.Println("Validation Completed")
		if err1 != nil {
			fmt.Println(err1)
			errorsUtils.SendErrorWithCustomMessage(w, err1.Error(), http.StatusUnauthorized)
			return
		}
		fmt.Println("Claims", claims)

		ctx := context.WithValue(r.Context(), constants.ClaimKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ValidateClientPermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(w http.ResponseWriter, r *http.Request) {
			err := recover()
			if err != nil {
				fmt.Println(err)
				errorsUtils.SendErrorWithCustomMessage(w, err.(error).Error(), http.StatusUnauthorized)
				return
			}
		}(w, r)
		fmt.Println("Admin Validation Middleware Called")
		fmt.Print(r.Context())
		claims := r.Context().Value("claims").(*encrypt.Claims)
		fmt.Print(claims)

		if claims.ClientId == 0 {
			errorsUtils.SendErrorWithCustomMessage(w, "Client Privileges Denied", http.StatusUnauthorized)
			return
		}

		requested_client_access, ok := mux.Vars(r)["client_id"]
		if !ok {
			errorsUtils.SendErrorWithCustomMessage(w, "Client Id not found. Please put Client Id in Path", http.StatusUnauthorized)
			return
		}
		requested_client_access_int, err := strconv.Atoi(requested_client_access)
		if err != nil {
			errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if requested_client_access_int != int(claims.ClientId) {
			errorsUtils.SendErrorWithCustomMessage(w, "You are not authorized to access this client", http.StatusUnauthorized)
			return
		}
		// ctx := context.WithValue(r.Context(), constants.ClaimsAdminKey, admin)

		next.ServeHTTP(w, r)
	})
}

func ValidateBankPermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(w http.ResponseWriter, r *http.Request) {
			err := recover()
			if err != nil {
				fmt.Println(err)
				errorsUtils.SendErrorWithCustomMessage(w, err.(error).Error(), http.StatusUnauthorized)
				return
			}
		}(w, r)
		fmt.Println("Bank Admin Validation Middleware Called")
		fmt.Print(r.Context())
		claims := r.Context().Value("claims").(*encrypt.Claims)
		fmt.Print(claims)

		if claims.BankId == 0 {
			errorsUtils.SendErrorWithCustomMessage(w, "Bank Privileges Denied", http.StatusUnauthorized)
			return
		}

		requested_bank_access, ok := mux.Vars(r)["bank_id"]
		if !ok {
			errorsUtils.SendErrorWithCustomMessage(w, "Bank Id not found. Please put Bank Id in Path", http.StatusUnauthorized)
			return
		}
		requested_bank_access_int, err := strconv.Atoi(requested_bank_access)
		if err != nil {
			errorsUtils.SendErrorWithCustomMessage(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if requested_bank_access_int != int(claims.BankId) {
			errorsUtils.SendErrorWithCustomMessage(w, "You are not authorized to access this client", http.StatusUnauthorized)
			return
		}
		// ctx := context.WithValue(r.Context(), constants.ClaimsAdminKey, admin)

		next.ServeHTTP(w, r)
	})
}

// func ValidateAdminPermissionsMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		defer func(w http.ResponseWriter, r *http.Request) {
// 			err := recover()
// 			if err != nil {
// 				fmt.Println(err)
// 				errorsUtils.SendErrorWithCustomMessage(w, err.(error).Error())
// 				return
// 			}
// 		}(w, r)
// 		fmt.Println("Admin Validation Middleware Called")
// 		fmt.Print(r.Context())
// 		claims := r.Context().Value("claims").(*encrypt.Claims)
// 		fmt.Print(claims.UserId)

// 		if err != nil {
// 			errorsUtils.SendErrorWithCustomMessage(w, err.Error())
// 			return
// 		}
// 		ctx := context.WithValue(r.Context(), constants.AdminKeyValue, admin)

// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
// func ValidateCustomerPermissionMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("Customer Validation Middleware Called")
// 		fmt.Print(r.Context())
// 		claims := r.Context().Value("claims").(*helper.Claims)
// 		fmt.Print(claims.UserId)
// 		customer, err := user.GetStaffInterfaceWithPassById(uint(claims.UserId))
// 		if err != nil {
// 			errorsUtils.SendErrorWithCustomMessage(w, err.Error())
// 			return
// 		}
// 		ctx := context.WithValue(r.Context(), constants.ClaimsCustomerKey, customer)

// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func getAuthTokenFromHeader(r *http.Request) (string, error) {
	headers := r.Header
	fmt.Println(headers)
	tempTokenHeader, ok := headers["Authorization"]
	if !ok || len(tempTokenHeader) == 0 {
		return "", errors.New("Token Not found")
	}
	token := tempTokenHeader[0]

	return token, nil
}
