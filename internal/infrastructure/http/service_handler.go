package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// ServiceHandler permet de creer les services tiers de l'ecosysteme.
//
// NOTE DE SECURITE : cette route (POST /services) n'est PAS encore
// protegee -- n'importe qui peut actuellement creer un service et obtenir
// une cle API. C'est un gap connu et deliberement laisse ouvert pour
// l'instant : il n'existe pas encore d'authentification usager (voir
// README, section Prochaines etapes). Une fois le mecanisme de session
// usager en place, cette route devra etre restreinte aux administrateurs
// uniquement.
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
