package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gcrg/kdb/service/pkg/db"
	"github.com/gcrg/kdb/service/pkg/middleware"
	"github.com/gcrg/kdb/service/pkg/models"
)

type SharesHandler struct {
	DB *db.DB
}

type CreateShareRequest struct {
	UsernameOrEmail string                 `json:"username_or_email"`
	Permission      models.SharePermission `json:"permission"`
}

type SharesListResponse struct {
	MyShares     []models.AccountShare `json:"my_shares"`     // People who have access to my account
	SharedWithMe []models.AccountShare `json:"shared_with_me"` // Accounts I have access to
}

func (h *SharesHandler) List(w http.ResponseWriter, r *http.Request) {
	authUser, _ := middleware.GetAuthUser(r.Context())

	// 1. People who have access to my account
	myRows, err := h.DB.Query(`
		SELECT s.id, s.owner_user_id, u_owner.username, s.shared_with_user_id, u_shared.username, s.permission, s.created_at
		FROM account_shares s
		JOIN users u_owner ON s.owner_user_id = u_owner.id
		JOIN users u_shared ON s.shared_with_user_id = u_shared.id
		WHERE s.owner_user_id = ?
		ORDER BY s.created_at DESC
	`, authUser.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch shared access records"}`, http.StatusInternalServerError)
		return
	}
	defer myRows.Close()

	myShares := make([]models.AccountShare, 0)
	for myRows.Next() {
		var s models.AccountShare
		if err := myRows.Scan(&s.ID, &s.OwnerUserID, &s.OwnerUsername, &s.SharedWithUserID, &s.SharedWithUsername, &s.Permission, &s.CreatedAt); err == nil {
			myShares = append(myShares, s)
		}
	}

	// 2. Accounts shared with me
	withMeRows, err := h.DB.Query(`
		SELECT s.id, s.owner_user_id, u_owner.username, s.shared_with_user_id, u_shared.username, s.permission, s.created_at
		FROM account_shares s
		JOIN users u_owner ON s.owner_user_id = u_owner.id
		JOIN users u_shared ON s.shared_with_user_id = u_shared.id
		WHERE s.shared_with_user_id = ?
		ORDER BY s.created_at DESC
	`, authUser.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch accounts shared with you"}`, http.StatusInternalServerError)
		return
	}
	defer withMeRows.Close()

	sharedWithMe := make([]models.AccountShare, 0)
	for withMeRows.Next() {
		var s models.AccountShare
		if err := withMeRows.Scan(&s.ID, &s.OwnerUserID, &s.OwnerUsername, &s.SharedWithUserID, &s.SharedWithUsername, &s.Permission, &s.CreatedAt); err == nil {
			sharedWithMe = append(sharedWithMe, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SharesListResponse{
		MyShares:     myShares,
		SharedWithMe: sharedWithMe,
	})
}

func (h *SharesHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUser, _ := middleware.GetAuthUser(r.Context())

	var req CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	searchTerm := strings.TrimSpace(req.UsernameOrEmail)
	if searchTerm == "" {
		http.Error(w, `{"error":"username or email is required"}`, http.StatusBadRequest)
		return
	}

	if req.Permission != models.SharePermissionRead && req.Permission != models.SharePermissionWrite {
		req.Permission = models.SharePermissionRead
	}

	// Find target user to share with
	var targetUser models.User
	err := h.DB.QueryRow(
		"SELECT id, username, email FROM users WHERE username = ? OR email = ?",
		searchTerm, strings.ToLower(searchTerm),
	).Scan(&targetUser.ID, &targetUser.Username, &targetUser.Email)

	if err != nil {
		http.Error(w, `{"error":"user not found with that username or email"}`, http.StatusNotFound)
		return
	}

	if targetUser.ID == authUser.ID {
		http.Error(w, `{"error":"you cannot share access with yourself"}`, http.StatusBadRequest)
		return
	}

	shareID := uuid.New().String()
	now := time.Now().UTC()

	// Insert or replace on conflict
	_, err = h.DB.Exec(`
		INSERT INTO account_shares (id, owner_user_id, shared_with_user_id, permission, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(owner_user_id, shared_with_user_id) DO UPDATE SET permission = excluded.permission
	`, shareID, authUser.ID, targetUser.ID, req.Permission, now)

	if err != nil {
		http.Error(w, `{"error":"failed to grant account access"}`, http.StatusInternalServerError)
		return
	}

	share := models.AccountShare{
		ID:                 shareID,
		OwnerUserID:        authUser.ID,
		OwnerUsername:      authUser.Username,
		SharedWithUserID:   targetUser.ID,
		SharedWithUsername: targetUser.Username,
		Permission:         req.Permission,
		CreatedAt:          now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(share)
}

func (h *SharesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authUser, _ := middleware.GetAuthUser(r.Context())
	shareID := chi.URLParam(r, "id")

	res, err := h.DB.Exec(
		"DELETE FROM account_shares WHERE id = ? AND (owner_user_id = ? OR shared_with_user_id = ?)",
		shareID, authUser.ID, authUser.ID,
	)
	if err != nil {
		http.Error(w, `{"error":"failed to revoke share"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error":"share record not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"access revoked successfully"}`))
}
