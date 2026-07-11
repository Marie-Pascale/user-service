package httpapi

import (
	"context"
	"errors"
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// contextKey est un type prive pour eviter les collisions avec d'autres
// packages qui pourraient aussi stocker des valeurs dans le context.Context
// (une chaine brute comme cle serait un risque de collision silencieuse).
type contextKey int

const serviceContextKey contextKey = iota

// ServiceAuthMiddleware authentifie les requetes provenant d'un AUTRE
// SERVICE de l'ecosysteme (pas d'un usager humain), via une cle API
// envoyee dans l'entete "X-API-Key". On utilise un entete distinct
// d'"Authorization" pour ne jamais entrer en conflit avec le futur
// mecanisme de session des usagers (qui utilisera "Authorization: Bearer").
type ServiceAuthMiddleware struct {
	Services application.ServiceRepository
}

func NewServiceAuthMiddleware(services application.ServiceRepository) *ServiceAuthMiddleware {
	return &ServiceAuthMiddleware{Services: services}
}

// Authenticate enveloppe un handler (ou un sous-mux) pour exiger une cle
// API valide. En cas de succes, le domain.Service authentifie est ajoute
// au contexte de la requete, recuperable via ServiceFromContext -- c'est
// ainsi qu'un handler protege sait QUEL service fait la demande, sans
// jamais faire confiance a un ID fourni dans l'URL ou le corps.
func (m *ServiceAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plainKey := r.Header.Get("X-API-Key")
		if plainKey == "" {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing X-API-Key header"})
			return
		}

		svc, err := m.Services.GetByAPIKeyHash(r.Context(), domain.HashAPIKey(plainKey))
		if err != nil {
			if errors.Is(err, application.ErrNotFound) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid API key"})
				return
			}
			writeError(w, err)
			return
		}

		// Double verification explicite, en plus de la recherche par hash --
		// redondant en pratique (une collision de hash SHA-256 est
		// virtuellement impossible), mais rend l'intention du code claire :
		// c'est CETTE ligne qui authentifie, pas seulement la requete SQL.
		if !svc.VerifyAPIKey(plainKey) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid API key"})
			return
		}

		ctx := context.WithValue(r.Context(), serviceContextKey, svc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ServiceFromContext recupere le service authentifie par le middleware.
// Le deuxieme retour est false si aucun service n'est present -- ce qui
// signale un handler protege appele sans passer par le middleware
// (erreur de cablage des routes, a corriger, pas une situation normale).
func ServiceFromContext(ctx context.Context) (*domain.Service, bool) {
	svc, ok := ctx.Value(serviceContextKey).(*domain.Service)
	return svc, ok
}
