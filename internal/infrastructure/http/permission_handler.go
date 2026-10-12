package httpapi

import (
	"errors"
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type PermissionHandler struct {
	Permissions application.PermissionRepository
}

func NewPermissionHandler(permissions application.PermissionRepository) *PermissionHandler {
	return &PermissionHandler{Permissions: permissions}
}

func (h *PermissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPermissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: invalidJSONRequestBody})
		return
	}

	p, err := domain.NewPermission(domain.NewID(), req.ServiceID, req.Key, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.Permissions.Create(r.Context(), p); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newPermissionResponse(p))
}

func (h *PermissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.Permissions.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newPermissionResponse(p))
}

func (h *PermissionHandler) RegisterSelf(w http.ResponseWriter, r *http.Request) {
	svc, ok := ServiceFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated service"})
		return
	}

	var req registerPermissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: invalidJSONRequestBody})
		return
	}

	existing, err := h.Permissions.GetByServiceIDAndKey(r.Context(), svc.ID, req.Key)
	if err == nil {
		writeJSON(w, http.StatusOK, newPermissionResponse(existing))
		return
	}
	if !errors.Is(err, application.ErrNotFound) {
		writeError(w, err)
		return
	}

	p, err := domain.NewPermission(domain.NewID(), svc.ID, req.Key, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.Permissions.Create(r.Context(), p); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newPermissionResponse(p))
}

func (h *PermissionHandler) ListByService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	if serviceID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "service_id query parameter is required"})
		return
	}

	permissions, err := h.Permissions.ListByServiceID(r.Context(), serviceID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]permissionResponse, len(permissions))
	for i, p := range permissions {
		resp[i] = newPermissionResponse(p)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *PermissionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Permissions.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
