package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
)

func BindRoutes(r *chi.Mux, app core.App) {
	authMiddleware := middleware.HttpAuthMiddleware(app)
	r.Use(authMiddleware)

	// r.Get("/api/ws", )
}
