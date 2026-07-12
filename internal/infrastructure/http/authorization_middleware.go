package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// AuthorizationMiddleware verifie qu'un usager authentifie (voir
// UserAuthMiddleware, qui DOIT s'executer avant celui-ci dans la chaine)
// possede la permission requise. Utilise le meme AuthorizationService que
// celui expose aux services tiers via GET /users/{id}/permissions, applique
// ici a ce service lui-meme (voir application.SelfServiceID).
type AuthorizationMiddleware struct {
	Authorization *application.AuthorizationService
}

func NewAuthorizationMiddleware(authorization *application.AuthorizationService) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{Authorization: authorization}
}

// RequirePermission exige que l'usager authentifie possede permissionKey,
// sans exception. Pour les routes qui ne concernent aucun usager en
// particulier (ex: POST /services, GET /users en liste complete).
//
// Retourne un "middleware factory" (func(http.Handler) http.Handler)
// plutot qu'un http.Handler directement -- meme forme que
// ServiceAuthMiddleware.Authenticate, pour rester composable de la
// meme facon dans router.go.
func (m *AuthorizationMiddleware) RequirePermission(permissionKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				// Ne devrait jamais arriver si les routes sont cablees
				// correctement -- UserAuthMiddleware doit toujours s'executer
				// avant (voir router.go).
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated"})
				return
			}

			allowed, err := m.Authorization.HasPermission(r.Context(), u.ID, application.SelfServiceID, permissionKey)
			if err != nil {
				writeError(w, err)
				return
			}
			if !allowed {
				writeJSON(w, http.StatusForbidden, errorResponse{Error: "insufficient permissions"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireSelfOrPermission autorise l'usager authentifie s'il agit sur SON
// PROPRE compte (le parametre de chemin "id" correspond a son propre ID),
// OU s'il possede permissionKey (ex: un admin qui gere n'importe quel
// usager). C'est la regle evoquee en concevant le domaine : "un usager X
// ne peut pas voir tous les usagers, mais peut voir son propre compte".
func (m *AuthorizationMiddleware) RequireSelfOrPermission(permissionKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated"})
				return
			}

			if r.PathValue("id") == u.ID {
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := m.Authorization.HasPermission(r.Context(), u.ID, application.SelfServiceID, permissionKey)
			if err != nil {
				writeError(w, err)
				return
			}
			if !allowed {
				writeJSON(w, http.StatusForbidden, errorResponse{Error: "insufficient permissions"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
