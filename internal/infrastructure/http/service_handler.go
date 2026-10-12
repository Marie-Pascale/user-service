package httpapi

import (
	"net/http"

	"user-service/internal/application"
	"user-service/internal/domain"
)

type ServiceHandler struct {
	Services application.ServiceRepository
}

func NewServiceHandler(services application.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{Services: services}
}

func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: invalidJSONRequestBody})
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

func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	svc, err := h.Services.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newServiceResponse(svc))
}

func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")

	var req createServiceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: invalidJSONRequestBody})
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

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.Services.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
