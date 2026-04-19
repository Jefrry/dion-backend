package router

import (
	"dion-backend/internal/handler"

	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	router *chi.Mux

	recordsHandler *handler.RecordsHandler
	artistsHandler *handler.ArtistsHandler
}

func NewRouter(rh *handler.RecordsHandler, ah *handler.ArtistsHandler) *Router {
	return &Router{
		recordsHandler: rh,
		artistsHandler: ah,
	}
}

func (r *Router) MustRun() http.Handler {
	router := chi.NewRouter()

	router.Route("/v1", func(v1 chi.Router) {
		v1.Get("/recordings", r.recordsHandler.GetApprovedList)
		v1.Get("/artists", r.artistsHandler.GetList)
	})

	r.router = router

	return router
}
