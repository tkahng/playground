package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
)

func BindRoutes(r *chi.Mux, app core.App) {
	authMiddleware := middleware.HttpAuthMiddleware(app)
	r.Use(authMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Route("/ws", func(r chi.Router) {
			// r.Get("/", app.statsHandler.SseHandler)
		})
	})
}
