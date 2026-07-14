package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// PermissionHandler regroupe les endpoints lies a Permission. Toutes ses
// routes sont protegees par la permission "permissions:manage" (voir
// router.go) -- il n'existe pas de notion de "soi-meme" pour une
// permission, contrairement a User.
type PermissionHandler struct {
	Permissions application.PermissionRepository
}

func NewPermissionHandler(permissions application.PermissionRepository) *PermissionHandler {
	return &PermissionHandler{Permissions: permissions}
}

// Create gere POST /permissions
func (h *PermissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPermissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
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

// Get gere GET /permissions/{id}
func (h *PermissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.Permissions.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newPermissionResponse(p))
}

// ListByService gere GET /permissions?service_id=...
// service_id est obligatoire -- lister TOUTES les permissions tous
// services confondus n'a pas d'usage clair pour l'instant (voir
// PermissionRepository, qui n'expose pas cette methode).
func (h *PermissionHandler) ListByService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	if serviceID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "le parametre de requete service_id est requis"})
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

// Delete gere DELETE /permissions/{id}
func (h *PermissionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Permissions.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
