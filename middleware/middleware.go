package middleware

import (
	"net/http"
)

func ValidateAdminPermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Header.Get("Role")

		if userRole != "SUPER_ADMIN" {
			http.Error(w, "Forbidden : SuperAdmin access only", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ValidateBankUserPermissionsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Header.Get("Role")

		if userRole != "BANK_USER" {
			http.Error(w, "Forbidden: Requires BANK_USER role", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
