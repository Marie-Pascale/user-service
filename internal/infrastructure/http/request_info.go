package httpapi

import (
	"net"
	"net/http"
	"strings"
)

// clientIP extrait l'adresse IP du client depuis la requete HTTP.
// Verifie d'abord X-Forwarded-For (pose par un reverse proxy en amont,
// ex: nginx, un load balancer) -- prend le PREMIER maillon de la chaine
// (l'IP originale du client, pas celle du proxy le plus proche). Se
// rabat sur RemoteAddr si l'entete est absent (connexion directe).
//
// NOTE DE SECURITE : X-Forwarded-For est un entete pose par le CLIENT,
// donc falsifiable par quiconque parle directement a ce serveur -- a ne
// faire confiance QUE si un reverse proxy de confiance est effectivement
// en amont et ecrase cet entete avant qu'il n'atteigne cette application.
// Ce n'est pas encore configure dans ce projet (deploiement direct, sans
// proxy devant l'API pour l'instant) -- voir le README.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
