package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gcrg/kdb/service/pkg/auth"
	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/models"
)

type contextKey string

const (
	UserKey             contextKey = "auth_user"
	TargetUserIDKey     contextKey = "target_user_id"
	TargetPermissionKey contextKey = "target_permission"
)

type AuthUser struct {
	ID       string
	Username string
	Email    string
}

// RequireAuth ensures that the request is authenticated with a valid JWT token.
func RequireAuth(jwtSecret []byte, database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			// 1. Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// 2. Check Cookie if header not found
			if tokenString == "" {
				if cookie, err := r.Cookie("kdb_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateToken(jwtSecret, tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			authUser := AuthUser{
				ID:       claims.UserID,
				Username: claims.Username,
				Email:    claims.Email,
			}

			targetUserID := authUser.ID
			targetPermission := models.SharePermissionWrite // Full access to own account

			// Check for X-Account-ID header (account delegation)
			requestedAccountID := r.Header.Get("X-Account-ID")
			if requestedAccountID != "" && requestedAccountID != authUser.ID {
				var permission string
				err := database.QueryRow(
					"SELECT permission FROM account_shares WHERE owner_user_id = ? AND shared_with_user_id = ?",
					requestedAccountID, authUser.ID,
				).Scan(&permission)

				if err != nil {
					http.Error(w, `{"error":"forbidden: no access to requested account"}`, http.StatusForbidden)
					return
				}

				targetUserID = requestedAccountID
				targetPermission = models.SharePermission(permission)

				// For modifying methods (POST, PUT, DELETE), enforce write permission
				if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
					if targetPermission != models.SharePermissionWrite {
						http.Error(w, `{"error":"forbidden: read-only access"}`, http.StatusForbidden)
						return
					}
				}
			}

			ctx := context.WithValue(r.Context(), UserKey, authUser)
			ctx = context.WithValue(ctx, TargetUserIDKey, targetUserID)
			ctx = context.WithValue(ctx, TargetPermissionKey, targetPermission)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthUser retrieves the authenticated user from the context.
func GetAuthUser(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(UserKey).(AuthUser)
	return u, ok
}

// GetTargetUserID retrieves the active account ID (personal or shared).
func GetTargetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(TargetUserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetTargetPermission retrieves the permission level on the target account.
func GetTargetPermission(ctx context.Context) models.SharePermission {
	if perm, ok := ctx.Value(TargetPermissionKey).(models.SharePermission); ok {
		return perm
	}
	return models.SharePermissionWrite
}
