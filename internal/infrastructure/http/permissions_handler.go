package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// PermissionsHandler expose aux AUTRES SERVICES de l'ecosysteme les
// permissions accordees a un usager -- c'est l'exigence originale du
// projet ("un systeme permettant aux autres services de venir chercher
// les permissions accordees a un usager en particulier, chacun identifie
// par son propre identifiant"). Protege par ServiceAuthMiddleware.
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

// ListForUser gere GET /users/{id}/permissions
//
// Point important : l'ID du service n'est JAMAIS lu depuis l'URL, le corps,
// ou un parametre de requete -- il vient EXCLUSIVEMENT du contexte, pose
// par le middleware apres verification de la cle API. Un service ne peut
// donc jamais usurper l'identite d'un autre service pour voir des
// permissions qui ne le concernent pas.
func (h *PermissionsHandler) ListForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	svc, ok := ServiceFromContext(r.Context())
	if !ok {
		// Ne devrait jamais arriver si les routes sont cablees correctement
		// (voir router.go) -- le middleware doit toujours s'executer avant.
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

// CheckForUser gere GET /users/{id}/permissions/check?permission=orders:read
//
// Raccourci pratique par rapport a ListForUser : plutot que de recuperer
// TOUTES les permissions puis chercher soi-meme si celle qui interesse le
// service y figure, un service peut demander directement "cet usager a-t-il
// CETTE permission precise ?" -- une seule cle a comparer, une reponse
// booleenne. Utile pour une verification synchrone avant d'executer une
// action (ex: avant de traiter une commande).
//
// Memes garanties de securite que ListForUser : service_id vient
// EXCLUSIVEMENT du contexte, jamais de la requete.
func (h *PermissionsHandler) CheckForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	svc, ok := ServiceFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated service"})
		return
	}

	permissionKey := r.URL.Query().Get("permission")
	if permissionKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "le parametre de requete permission est requis"})
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
