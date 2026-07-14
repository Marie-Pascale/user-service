package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// GroupHandler regroupe les endpoints lies a Group. Toutes ses routes
// sont protegees par la permission "groups:manage".
type GroupHandler struct {
	Groups application.GroupRepository
}

func NewGroupHandler(groups application.GroupRepository) *GroupHandler {
	return &GroupHandler{Groups: groups}
}

// Create gere POST /groups
func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createGroupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	g, err := domain.NewGroup(domain.NewID(), req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.Groups.Create(r.Context(), g); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newGroupResponse(g))
}

// Get gere GET /groups/{id}
func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	g, err := h.Groups.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newGroupResponse(g))
}

// List gere GET /groups
func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	groups, err := h.Groups.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]groupResponse, len(groups))
	for i, g := range groups {
		resp[i] = newGroupResponse(g)
	}
	writeJSON(w, http.StatusOK, resp)
}

// AddMember gere POST /groups/{id}/members
func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")

	var req groupAddMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	if err := h.Groups.AddMember(r.Context(), groupID, req.UserID); err != nil {
		writeError(w, err)
		return
	}

	g, err := h.Groups.GetByID(r.Context(), groupID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newGroupResponse(g))
}

// RemoveMember gere DELETE /groups/{id}/members/{userId}
func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")
	userID := r.PathValue("userId")

	if err := h.Groups.RemoveMember(r.Context(), groupID, userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AssignRole gere POST /groups/{id}/roles
func (h *GroupHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")

	var req groupAssignRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	if err := h.Groups.AssignRole(r.Context(), groupID, req.RoleID); err != nil {
		writeError(w, err)
		return
	}

	g, err := h.Groups.GetByID(r.Context(), groupID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newGroupResponse(g))
}

// Delete gere DELETE /groups/{id}
func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Groups.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
