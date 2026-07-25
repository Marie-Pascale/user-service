package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// ServiceHandler gere les endpoints lies a Service. La creation (Create)
// est protegee par la permission "services:create" ; les autres
// operations (List, Get, Delete, RotateAPIKey) par "services:manage" --
// voir router.go et la doc de application.PermServicesManage pour la
// raison de cette separation.
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
		UpdatedAt:   svc.UpdatedAt,
	})
}

// Get gere GET /services/{id}. Ne retourne jamais APIKeyHash, meme hache
// (voir serviceResponse).
func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	svc, err := h.Services.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newServiceResponse(svc))
}

// Update gere PUT /services/{id} -- modifie Name/Description uniquement.
// RotateAPIKey reste la seule facon de changer la cle API.
func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")

	var req createServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	svc, err := h.Services.GetByID(r.Context(), serviceID)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := svc.UpdateDetails(req.Name, req.Description); err != nil {
		writeError(w, err)
		return
	}

	if err := h.Services.Update(r.Context(), svc); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newServiceResponse(svc))
}

// List gere GET /services
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	services, err := h.Services.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]serviceResponse, len(services))
	for i, svc := range services {
		resp[i] = newServiceResponse(svc)
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete gere DELETE /services/{id}
func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Services.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RotateAPIKey gere POST /services/{id}/rotate-key. Genere une nouvelle
// cle API pour ce service, invalidant l'ancienne immediatement -- utile
// en cas de compromission suspectee ou de rotation planifiee. Comme a la
// creation, la nouvelle cle en clair n'apparait que dans CETTE reponse.
func (h *ServiceHandler) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	svc, err := h.Services.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	plainAPIKey, err := svc.RotateAPIKey()
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.Services.UpdateAPIKeyHash(r.Context(), svc.ID, svc.APIKeyHash); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, createServiceResponse{
		ID:          svc.ID,
		Name:        svc.Name,
		Description: svc.Description,
		APIKey:      plainAPIKey,
		UpdatedAt:   svc.UpdatedAt,
	})
}
