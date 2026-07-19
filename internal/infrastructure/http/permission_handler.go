package httpapi

import (
	"errors"
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// PermissionHandler regroupe les endpoints lies a Permission. La plupart
// de ses routes sont protegees par la permission "permissions:manage"
// (session usager) -- sauf RegisterSelf, protegee par cle API (voir
// router.go pour le detail de chaque route).
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

// RegisterSelf gere POST /permissions/self -- protegee par
// ServiceAuthMiddleware (cle API), PAS par une session usager. Permet a
// un service tiers de l'ecosysteme de declarer lui-meme ses propres
// permissions, sans intervention d'un admin humain -- utile pour un
// enregistrement automatique au demarrage du service.
//
// Idempotente : si une permission avec la meme cle existe deja pour ce
// service, elle est retournee telle quelle (200) plutot que de tenter une
// creation en double (qui echouerait sur la contrainte UNIQUE(service_id,
// key) en base). Un service qui s'enregistre a chaque redemarrage n'a
// donc pas a se soucier de l'etat existant.
//
// Comme pour PermissionsHandler.ListForUser, service_id vient
// EXCLUSIVEMENT du contexte (pose par le middleware apres verification de
// la cle API) -- jamais du corps de la requete. Un service ne peut donc
// jamais enregistrer une permission au nom d'un AUTRE service.
func (h *PermissionHandler) RegisterSelf(w http.ResponseWriter, r *http.Request) {
	svc, ok := ServiceFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated service"})
		return
	}

	var req registerPermissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
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
