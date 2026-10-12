package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

type AuthorizationMiddleware struct {
	Authorization *application.AuthorizationService
}

func NewAuthorizationMiddleware(authorization *application.AuthorizationService) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{Authorization: authorization}
}

func (m *AuthorizationMiddleware) RequirePermission(permissionKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
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
