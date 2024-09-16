package handlers

import (
	"encoding/json"
	"net/http"
)

// ListSchemas godoc
//
//	@Summary		List schemas from Confluent Cloud
//	@Description	Get a list of schemas from the Confluent Cloud API.
//	@Tags			schemas
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		models.Schema
//	@Failure		500	{object}	ErrorResponse
//	@Router			/schemas [get]
//
//	@Param			subjectPrefix	query	string	false	"Subject prefix to filter schemas by"
func ListSchemas(h *Handler, w http.ResponseWriter, r *http.Request, subjectPrefix string) {

	schemas, err := h.SchemaService.ListSchemas(h.Ctx, subjectPrefix)

	if err != nil {
		h.Logger.Error(err, "failed to list schemas")
		http.Error(w, "Failed to list schemas", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schemas)
}
