package handlers

import (
	"encoding/json"
	"net/http"
)

// Health godoc
//
//	@Summary		Health endpoint
//	@Description	Returns health status of the API
//	@Tags			health
//	@Accept			json
//	@Produce		json
//
//	@Success		200	{object}	map[string]string
//
//	@Router			/health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "ok"}
	json.NewEncoder(w).Encode(response)
}
