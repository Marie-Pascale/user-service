package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

type PermissionsHandler struct {
	Authorization *application.AuthorizationService
}

func NewPermissionsHandler(authorization *application.AuthorizationService) *PermissionsHandler {
	return &PermissionsHandler{Authorization: authorization}
}

type permissionsResponse struct {
	UserID      string   `json:"user_id"`
	ServiceID   string   `json:"service_id"`
	Permissions []string `json:"permissions"`
}

type checkPermissionResponse struct {
	UserID     string `json:"user_id"`
	ServiceID  string `json:"service_id"`
	Permission string `json:"permission"`
	Allowed    bool   `json:"allowed"`
}

func (h *PermissionsHandler) ListForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	svc, ok := ServiceFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated service"})
		return
	}

	keys, err := h.Authorization.ResolvePermissionKeys(r.Context(), userID, svc.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, permissionsResponse{
		UserID:      userID,
		ServiceID:   svc.ID,
		Permissions: keys,
	})
}

func (h *PermissionsHandler) CheckForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	svc, ok := ServiceFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated service"})
		return
	}

	permissionKey := r.URL.Query().Get("permission")
	if permissionKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "permission query parameter is required"})
		return
	}

	allowed, err := h.Authorization.HasPermission(r.Context(), userID, svc.ID, permissionKey)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, checkPermissionResponse{
		UserID:     userID,
		ServiceID:  svc.ID,
		Permission: permissionKey,
		Allowed:    allowed,
	})
}
