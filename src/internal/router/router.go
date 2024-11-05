package router

import (
	"net/http"

	"github.com/dfds/confluent-gateway/internal/handlers"
	"github.com/dfds/confluent-gateway/internal/models"
	httpSwagger "github.com/swaggo/http-swagger"
)

//	@title			Confluent Gateway API
//	@version		0.1
//	@description	This is a DFDS REST API for interacting with the Confluent Cloud API

// @contact.name	Cloud Engineering
// @contact.email	cloud.engineering@dfds.com
func SetupRoutes(handler *handlers.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.Health(w, r)
	})

	mux.Handle("/swagger", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /clusters/{clusterId}/schemas", func(w http.ResponseWriter, r *http.Request) {
		clusterId := models.ClusterId(r.PathValue("clusterId"))

		subjectPrefix := r.URL.Query().Get("subjectPrefix")

		handlers.ListSchemas(handler, w, r, subjectPrefix, clusterId)
	})

	return mux
}
