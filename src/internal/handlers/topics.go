package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dfds/confluent-gateway/internal/models"
)

// ListSchemas godoc
//
//	@Summary		List topics from Confluent Cloud
//	@Description	Get a list of topics from the Confluent Cloud API.
//	@Tags			topics
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		models.ConfluentTopic
//	@Failure		500	{object}	ErrorResponse
//	@Router			/topics [get]
func ListTopics(h *Handler, w http.ResponseWriter, r *http.Request, clusterId models.ClusterId) {

	topics, err := h.TopicService.ListTopics(h.Ctx, clusterId)

	if err != nil {
		h.Logger.Error(err, "failed to list topics")
		http.Error(w, "Failed to list topics", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(topics)
}
