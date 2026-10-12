package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type UserAuthMiddleware struct {
	Sessions application.SessionRepository
	Users    application.UserRepository
}

func NewUserAuthMiddleware(sessions application.SessionRepository, users application.UserRepository) *UserAuthMiddleware {
	return &UserAuthMiddleware{Sessions: sessions, Users: users}
}

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
			_ = m.Sessions.Delete(r.Context(), session.ID)
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "session expired"})
			return
		}

		if session.IsSuspicious(clientIP(r), r.Header.Get("User-Agent")) {
			_ = m.Sessions.Delete(r.Context(), session.ID)
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "session revoked: suspicious activity detected"})
			return
		}

		user, err := m.Users.GetByID(r.Context(), session.UserID)
		if err != nil {
			writeError(w, err)
			return
		}

		_ = m.Sessions.Touch(r.Context(), session.ID)

		ctx := context.WithValue(r.Context(), userContextKey, user)
		ctx = context.WithValue(ctx, sessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	u, ok := ctx.Value(userContextKey).(*domain.User)
	return u, ok
}

func SessionFromContext(ctx context.Context) (*domain.Session, bool) {
	s, ok := ctx.Value(sessionContextKey).(*domain.Session)
	return s, ok
}
