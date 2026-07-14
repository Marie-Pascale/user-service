package httpapi

import "net/http"

// CORSMiddleware gere les entetes Cross-Origin Resource Sharing, pour
// permettre a un frontend servi depuis une autre origine (ex: un serveur
// de dev Vite/React sur localhost:5173) de consommer cette API depuis un
// navigateur.
//
// Contrairement aux autres middlewares du projet (ServiceAuthMiddleware,
// UserAuthMiddleware, AuthorizationMiddleware), celui-ci enveloppe TOUT
// le routeur plutot qu'une route precise -- un navigateur peut envoyer
// une requete preflight (OPTIONS) vers N'IMPORTE QUELLE route, et les
// entetes CORS doivent etre presents sur TOUTES les reponses (succes ET
// erreurs). Voir Wrap et son usage dans cmd/api/main.go.
//
// Note : ce projet authentifie via un entete "Authorization"/"X-API-Key"
// explicitement pose par le code JS du frontend, pas via des cookies
// geres automatiquement par le navigateur -- Access-Control-Allow-Credentials
// n'est donc pas necessaire ici (il ne sert qu'au mode "credentials:
// include" des cookies cross-origin).
type CORSMiddleware struct {
	// AllowedOrigins liste les origines autorisees (ex:
	// "http://localhost:5173"). La valeur speciale "*" autorise
	// n'importe quelle origine -- pratique en developpement, mais a
	// restreindre explicitement des que l'API est exposee publiquement.
	AllowedOrigins []string
}

func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{AllowedOrigins: allowedOrigins}
}

// Wrap enveloppe un handler (typiquement le *http.ServeMux entier) pour y
// ajouter les entetes CORS et repondre directement aux requetes preflight
// (OPTIONS), sans jamais les laisser atteindre les handlers applicatifs.
func (m *CORSMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin != "" && m.isAllowed(origin) {
			if m.hasWildcard() {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// On reflete l'origine precise (plutot que "*") quand la
				// liste est restreinte -- necessaire pour que le navigateur
				// accepte la reponse, et plus sur qu'un wildcard global.
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		}

		// Une requete preflight ne va jamais plus loin qu'ici -- le
		// navigateur veut seulement savoir si la vraie requete sera
		// autorisee, il n'attend aucun corps de reponse.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *CORSMiddleware) isAllowed(origin string) bool {
	for _, allowed := range m.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (m *CORSMiddleware) hasWildcard() bool {
	for _, allowed := range m.AllowedOrigins {
		if allowed == "*" {
			return true
		}
	}
	return false
}
