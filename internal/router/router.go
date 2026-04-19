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
		v1.Route("/recordings", func(rec chi.Router) {
			rec.Get("/", r.recordsHandler.GetApprovedList)
			rec.Get("/{slug}", r.recordsHandler.GetBySlug)
		})

		v1.Route("/artists", func(art chi.Router) {
			art.Get("/", r.artistsHandler.GetList)
			art.Get("/{slug}", r.artistsHandler.GetBySlug)
			art.Get("/{slug}/recordings", r.recordsHandler.GetListByArtistSlug)
		})
	})

	r.router = router

	return router
}
