package router

import (
	"dion-backend/internal/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
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

	router.Use(corsMiddleware)

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/v1", func(v1 chi.Router) {
		v1.Route("/recordings", func(rec chi.Router) {
			rec.Get("/", r.recordsHandler.GetApprovedList)
			rec.Post("/", r.recordsHandler.Create)
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

// TODO: Remove it before prod
func corsMiddleware(next http.Handler) http.Handler {
	const allowedOrigin = "http://localhost:8082"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
