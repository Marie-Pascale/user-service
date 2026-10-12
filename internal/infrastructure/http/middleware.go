package httpapi

import (
	"context"
	"errors"
	"net/http"

	"user-service/internal/application"
	"user-service/internal/domain"
)

type contextKey int

const (
	serviceContextKey contextKey = iota
	userContextKey
	sessionContextKey
)

type ServiceAuthMiddleware struct {
	Services application.ServiceRepository
}

func NewServiceAuthMiddleware(services application.ServiceRepository) *ServiceAuthMiddleware {
	return &ServiceAuthMiddleware{Services: services}
}

func (m *ServiceAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plainKey := r.Header.Get("X-API-Key")
		if plainKey == "" {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing X-API-Key header"})
			return
		}

		svc, err := m.Services.GetByAPIKeyHash(r.Context(), domain.HashToken(plainKey))
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid API key"})
				return
			}
			writeError(w, err)
			return
		}

		if !svc.VerifyAPIKey(plainKey) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid API key"})
			return
		}

		ctx := context.WithValue(r.Context(), serviceContextKey, svc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ServiceFromContext(ctx context.Context) (*domain.Service, bool) {
	svc, ok := ctx.Value(serviceContextKey).(*domain.Service)
	return svc, ok
}
