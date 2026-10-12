package httpapi

import (
	"errors"
	"log"
	"net/http"

	"user-service/internal/application"
	"user-service/internal/domain"
)

const invalidJSONRequestBody = "invalid JSON request body"

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
		log.Printf("erreur interne non geree: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

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
