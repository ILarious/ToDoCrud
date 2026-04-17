package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"todo_crud/internal/app/http/handler"
)

func NewRouter(authHandler *handler.AuthHandler, listHandler *handler.ListHandler, itemHandler *handler.ItemHandler, authService TokenParser) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(CORSMiddleware)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/sign-up", authHandler.SignUp)
		r.Post("/sign-in", authHandler.SignIn)
	})

	r.Route("/api/v1/lists", func(r chi.Router) {
		r.Use(AuthMiddleware(authService))

		r.Get("/", listHandler.List)
		r.Post("/", listHandler.Create)

		r.Route("/{listId}", func(r chi.Router) {
			r.Get("/", listHandler.GetByID)
			r.Patch("/", listHandler.Update)
			r.Delete("/", listHandler.Delete)

			r.Route("/items", func(r chi.Router) {
				r.Get("/", itemHandler.List)
				r.Post("/", itemHandler.Create)

				r.Route("/{itemId}", func(r chi.Router) {
					r.Get("/", itemHandler.GetByID)
					r.Patch("/", itemHandler.Update)
					r.Delete("/", itemHandler.Delete)
				})
			})
		})
	})

	return r
}
