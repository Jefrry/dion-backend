package router

import (
	"dion-backend/internal/config"
	"dion-backend/internal/handler"
	"dion-backend/internal/middlewares"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

const swaggerBearerRequestInterceptor = `(request) => {
  const authorization = request.headers.Authorization || request.headers.authorization;
  if (authorization && !/^Bearer\s+/i.test(authorization)) {
    request.headers.Authorization = "Bearer " + authorization;
    delete request.headers.authorization;
  }
  return request;
}`

type Router struct {
	router *chi.Mux

	recordsHandler *handler.RecordsHandler
	artistsHandler *handler.ArtistsHandler
	adminHandler   *handler.AdminHandler
	adminConfig    config.AdminConfig
}

func NewRouter(rh *handler.RecordsHandler, ah *handler.ArtistsHandler, adminHandler *handler.AdminHandler, adminConfig config.AdminConfig) *Router {
	return &Router{
		recordsHandler: rh,
		artistsHandler: ah,
		adminHandler:   adminHandler,
		adminConfig:    adminConfig,
	}
}

func (r *Router) MustRun() http.Handler {
	router := chi.NewRouter()
	adminLoginRateLimiter := middlewares.NewAdminLoginRateLimiter()

	router.Use(middlewares.CORS)

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
		httpSwagger.UIConfig(map[string]string{
			"requestInterceptor": swaggerBearerRequestInterceptor,
			"showMutatedRequest": "true",
		}),
	))

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

		v1.Route("/admin", func(admin chi.Router) {
			admin.Use(middlewares.AdminAuth(r.adminConfig))
			admin.With(adminLoginRateLimiter).Post("/login", r.adminHandler.Login)
			admin.Get("/recordings/pending", r.recordsHandler.GetPendingList)
			admin.Patch("/recordings/{id}", r.recordsHandler.Update)
		})
	})

	r.router = router

	return router
}
