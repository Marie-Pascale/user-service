package httpapi

import (
	"errors"
	"log"
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// writeError traduit une erreur du domaine ou de la couche application en
// reponse HTTP appropriee. C'est le SEUL endroit ou cette traduction a
// lieu -- le domaine et l'application ne connaissent aucun code de statut
// HTTP, seulement des erreurs Go ordinaires (voir domain/user.go,
// application/errors.go).
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "resource not found"})

	case errors.Is(err, application.ErrEmailAlreadyUsed),
		errors.Is(err, application.ErrProviderAccountAlreadyLinked):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})

	case errors.Is(err, application.ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: err.Error()})

	case isValidationError(err):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})

	default:
		// Une erreur non reconnue (ex: la base de donnees est injoignable)
		// ne doit jamais fuiter de details techniques au client -- on la
		// journalise cote serveur, et on renvoie un message generique.
		log.Printf("erreur interne non geree: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

// isValidationError regroupe toutes les erreurs de validation exposees par
// le domaine -- celles qui doivent devenir un 400 Bad Request plutot qu'un
// 500. A completer au fur et a mesure qu'on ajoute des entites/handlers.
func isValidationError(err error) bool {
	return errors.Is(err, domain.ErrEmptyName) ||
		errors.Is(err, domain.ErrInvalidEmail) ||
		errors.Is(err, domain.ErrEmptyPassword) ||
		errors.Is(err, domain.ErrEmptyUserID) ||
		errors.Is(err, domain.ErrInvalidProvider) ||
		errors.Is(err, domain.ErrEmptyProviderUserID) ||
		errors.Is(err, domain.ErrEmptyServiceID) ||
		errors.Is(err, domain.ErrEmptyServiceName) ||
		errors.Is(err, domain.ErrEmptyPermissionKey) ||
		errors.Is(err, domain.ErrEmptyRoleName) ||
		errors.Is(err, domain.ErrEmptyGroupName)
}
