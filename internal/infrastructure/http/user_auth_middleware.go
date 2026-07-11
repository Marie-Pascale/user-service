package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// UserAuthMiddleware authentifie les requetes provenant d'un USAGER humain
// (par opposition a ServiceAuthMiddleware, pour les services tiers), via
// un jeton de session envoye dans l'entete standard "Authorization: Bearer
// <jeton>". Le jeton est obtenu via POST /auth/login.
type UserAuthMiddleware struct {
	Sessions application.SessionRepository
	Users    application.UserRepository
}

func NewUserAuthMiddleware(sessions application.SessionRepository, users application.UserRepository) *UserAuthMiddleware {
	return &UserAuthMiddleware{Sessions: sessions, Users: users}
}

// Authenticate enveloppe un handler pour exiger une session valide et non
// expiree. En cas de succes, l'usager ET sa session sont ajoutes au
// contexte -- recuperables via UserFromContext et SessionFromContext.
func (m *UserAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plainToken, ok := extractBearerToken(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing or malformed Authorization header"})
			return
		}

		session, err := m.Sessions.GetByTokenHash(r.Context(), domain.HashToken(plainToken))
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid session"})
				return
			}
			writeError(w, err)
			return
		}

		if !session.VerifyToken(plainToken) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid session"})
			return
		}

		if session.IsExpired() {
			// Nettoyage opportuniste : une session expiree ne sert plus a
			// rien, autant la supprimer maintenant plutot que d'attendre
			// une purge periodique separee. On ignore l'erreur -- ce n'est
			// pas grave si ca echoue, une prochaine requete la retentera.
			_ = m.Sessions.Delete(r.Context(), session.ID)
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "session expired"})
			return
		}

		user, err := m.Users.GetByID(r.Context(), session.UserID)
		if err != nil {
			writeError(w, err)
			return
		}

		// Best-effort : on ne bloque pas la requete si la mise a jour de
		// LastUsedAt echoue, ce n'est pas critique.
		_ = m.Sessions.Touch(r.Context(), session.ID)

		ctx := context.WithValue(r.Context(), userContextKey, user)
		ctx = context.WithValue(ctx, sessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractBearerToken lit le jeton depuis "Authorization: Bearer <jeton>".
// Retourne ok=false si l'entete est absent ou mal forme.
func extractBearerToken(r *http.Request) (token string, ok bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token = strings.TrimPrefix(header, prefix)
	if token == "" {
		return "", false
	}
	return token, true
}

// UserFromContext recupere l'usager authentifie par le middleware.
func UserFromContext(ctx context.Context) (*domain.User, bool) {
	u, ok := ctx.Value(userContextKey).(*domain.User)
	return u, ok
}

// SessionFromContext recupere la session authentifiee par le middleware --
// utile notamment pour Logout, qui a besoin de l'ID de session a revoquer.
func SessionFromContext(ctx context.Context) (*domain.Session, bool) {
	s, ok := ctx.Value(sessionContextKey).(*domain.Session)
	return s, ok
}
