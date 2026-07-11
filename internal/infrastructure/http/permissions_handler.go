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
