package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gcrg/kdb/service/pkg/auth"
	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

type AuthHandler struct {
	DB        *db.DB
	JWTSecret []byte
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if len(req.Username) < 3 || len(req.Password) < 6 || !strings.Contains(req.Email, "@") {
		http.Error(w, `{"error":"username must be >= 3 chars, password >= 6 chars, and a valid email"}`, http.StatusBadRequest)
		return
	}

	// Check if username or email already exists
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", req.Username, req.Email).Scan(&count)
	if err != nil {
		http.Error(w, `{"error":"internal database error"}`, http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, `{"error":"username or email already in use"}`, http.StatusConflict)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	var userCount int
	_ = h.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	isAdmin := userCount == 0 // First registered user is automatically admin
	if adminList := os.Getenv("KDB_ADMIN_USERS"); adminList != "" {
		for _, u := range strings.Split(adminList, ",") {
			trimmed := strings.TrimSpace(u)
			if strings.EqualFold(trimmed, req.Username) || strings.EqualFold(trimmed, req.Email) {
				isAdmin = true
				break
			}
		}
	}

	user := models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		IsAdmin:      isAdmin,
		CreatedAt:    time.Now().UTC(),
	}

	_, err = h.DB.Exec(
		"INSERT INTO users (id, username, email, password_hash, is_admin, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		user.ID, user.Username, user.Email, user.PasswordHash, user.IsAdmin, user.CreatedAt,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(h.JWTSecret, user.ID, user.Username, user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, `{"error":"failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	h.setAuthCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	queryTerm := strings.TrimSpace(req.UsernameOrEmail)
	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, username, email, password_hash, is_admin, created_at FROM users WHERE username = ? OR email = ?",
		queryTerm, strings.ToLower(queryTerm),
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.CreatedAt)

	if err != nil || !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error":"invalid username/email or password"}`, http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(h.JWTSecret, user.ID, user.Username, user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, `{"error":"failed to generate session token"}`, http.StatusInternalServerError)
		return
	}

	h.setAuthCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "kdb_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"logged out successfully"}`))
}

type SharedAccountInfo struct {
	OwnerID    string                 `json:"owner_id"`
	Username   string                 `json:"username"`
	Permission models.SharePermission `json:"permission"`
}

type MeResponse struct {
	User           models.User         `json:"user"`
	HasRoutes      bool                `json:"has_routes"`
	HasHome        bool                `json:"has_home"`
	SharedAccounts []SharedAccountInfo `json:"shared_accounts"`
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, username, email, is_admin, created_at FROM users WHERE id = ?",
		authUser.ID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	// Check if user has routes
	var routeCount int
	_ = h.DB.QueryRow("SELECT COUNT(*) FROM routes WHERE user_id = ?", user.ID).Scan(&routeCount)

	// Check if user has home location
	var homeCount int
	_ = h.DB.QueryRow("SELECT COUNT(*) FROM locations WHERE user_id = ? AND type = 'home'", user.ID).Scan(&homeCount)

	// Get shared accounts
	rows, err := h.DB.Query(`
		SELECT s.owner_user_id, u.username, s.permission
		FROM account_shares s
		JOIN users u ON s.owner_user_id = u.id
		WHERE s.shared_with_user_id = ?
	`, user.ID)

	sharedAccounts := make([]SharedAccountInfo, 0)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var acc SharedAccountInfo
			if err := rows.Scan(&acc.OwnerID, &acc.Username, &acc.Permission); err == nil {
				sharedAccounts = append(sharedAccounts, acc)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MeResponse{
		User:           user,
		HasRoutes:      routeCount > 0,
		HasHome:        homeCount > 0,
		SharedAccounts: sharedAccounts,
	})
}

func (h *AuthHandler) setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "kdb_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if len(req.Username) < 3 || !strings.Contains(req.Email, "@") {
		http.Error(w, `{"error":"username must be >= 3 chars and email must be valid"}`, http.StatusBadRequest)
		return
	}

	// Check if username or email is already taken by another user
	var count int
	err := h.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE (username = ? OR email = ?) AND id != ?",
		req.Username, req.Email, authUser.ID,
	).Scan(&count)
	if err != nil {
		http.Error(w, `{"error":"internal database error"}`, http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, `{"error":"username or email already in use"}`, http.StatusConflict)
		return
	}

	_, err = h.DB.Exec(
		"UPDATE users SET username = ?, email = ? WHERE id = ?",
		req.Username, req.Email, authUser.ID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to update profile"}`, http.StatusInternalServerError)
		return
	}

	// Fetch updated user
	var user models.User
	_ = h.DB.QueryRow(
		"SELECT id, username, email, is_admin, created_at FROM users WHERE id = ?",
		authUser.ID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin, &user.CreatedAt)

	// Issue updated JWT token
	token, err := auth.GenerateToken(h.JWTSecret, user.ID, user.Username, user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, `{"error":"failed to generate updated session token"}`, http.StatusInternalServerError)
		return
	}

	h.setAuthCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	authUser, ok := middleware.GetAuthUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 6 {
		http.Error(w, `{"error":"new password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	var currentHash string
	err := h.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", authUser.ID).Scan(&currentHash)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	if !auth.CheckPassword(req.CurrentPassword, currentHash) {
		http.Error(w, `{"error":"current password is incorrect"}`, http.StatusUnauthorized)
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, `{"error":"failed to hash new password"}`, http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", newHash, authUser.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to update password"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"password changed successfully"}`))
}
