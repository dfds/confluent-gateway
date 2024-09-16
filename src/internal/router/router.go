package router

import (
	"net/http"

	"github.com/dfds/confluent-gateway/internal/handlers"
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

	mux.HandleFunc("GET /schemas", func(w http.ResponseWriter, r *http.Request) {

		subjectPrefix := r.URL.Query().Get("subjectPrefix")
		handlers.ListSchemas(handler, w, r, subjectPrefix)
	})

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return mux
}
