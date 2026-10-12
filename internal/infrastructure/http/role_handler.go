package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type RoleHandler struct {
	Roles application.RoleRepository
}

func NewRoleHandler(roles application.RoleRepository) *RoleHandler {
	return &RoleHandler{Roles: roles}
}

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	role, err := domain.NewRole(domain.NewID(), req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.Roles.Create(r.Context(), role); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newRoleResponse(role))
}

func (h *RoleHandler) Get(w http.ResponseWriter, r *http.Request) {
	role, err := h.Roles.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newRoleResponse(role))
}

func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")

	var req createRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	role, err := h.Roles.GetByID(r.Context(), roleID)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := role.UpdateDetails(req.Name, req.Description); err != nil {
		writeError(w, err)
		return
	}

	if err := h.Roles.Update(r.Context(), role); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRoleResponse(role))
}

func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	roles, err := h.Roles.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]roleResponse, len(roles))
	for i, role := range roles {
		resp[i] = newRoleResponse(role)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *RoleHandler) GrantPermission(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")

	var req roleGrantPermissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	if err := h.Roles.GrantPermission(r.Context(), roleID, req.PermissionID); err != nil {
		writeError(w, err)
		return
	}

	role, err := h.Roles.GetByID(r.Context(), roleID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newRoleResponse(role))
}

func (h *RoleHandler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	permissionID := r.PathValue("permissionId")

	if err := h.Roles.RevokePermission(r.Context(), roleID, permissionID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Roles.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
