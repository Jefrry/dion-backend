package router

import (
	"dion-backend/internal/handler"

	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	router *chi.Mux

	recordsHandler *handler.RecordsHandler
}

func NewRouter(rh *handler.RecordsHandler) *Router {
	return &Router{
		recordsHandler: rh,
	}
}

func (r *Router) MustRun() http.Handler {
	router := chi.NewRouter()

	router.Route("/v1", func(v1 chi.Router) {
		v1.Get("/records", r.recordsHandler.Get)
	})

	r.router = router

	return router
}
