package handler

import (
	"net/http"

	"github.com/dvnpv/subscriptions-aggregator/internal/httpresponse"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
