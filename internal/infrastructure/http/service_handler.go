package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// ServiceHandler permet de creer les services tiers de l'ecosysteme.
// Protegee par authz.RequirePermission(application.PermServicesCreate)
// (voir router.go) -- seul un usager avec cette permission (typiquement
// via le role "admin") peut creer un service et obtenir sa cle API.
type ServiceHandler struct {
	Services application.ServiceRepository
}

func NewServiceHandler(services application.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{Services: services}
}

// Create gere POST /services. La cle API en clair n'apparait que dans
// CETTE reponse -- elle ne sera plus jamais recuperable ensuite (voir
// domain.NewService).
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	svc, plainAPIKey, err := domain.NewService(domain.NewID(), req.Name, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.Services.Create(r.Context(), svc); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createServiceResponse{
		ID:          svc.ID,
		Name:        svc.Name,
		Description: svc.Description,
		APIKey:      plainAPIKey,
	})
}
