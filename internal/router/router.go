// Package router настраивает маршрутизацию HTTP‑запросов к соответствующим обработчикам.
package router

import (
	"net/http"

	"github.com/amfib87/go_musthave-GophKeeper/internal/handlers"
	"github.com/go-chi/chi"
)

type Router struct {
	chi *chi.Mux
}

// Initialize создаёт и конфигурирует HTTP‑роутер с заданными обработчиками.
// Параметры:
//   - handler (*handlers.Handler): экземпляр HTTP‑обработчика.
//
// Возвращает:
//   - Router: готовый маршрутизатор для передачи в http.ListenAndServe;
//   - error: ошибка конфигурации (если есть).
func Initialize(h *handlers.Handler) (*Router, error) {
	r := &Router{chi: chi.NewRouter()}

	r.chi.Post("/api/register", h.Register)
	r.chi.Post("/api/login", h.Login)

	r.chi.Route("/api", func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		r.Get("/data", h.GetUserData)
		r.Post("/data", h.CreateData)
		r.Put("/data/", h.UpdateData)
		r.Delete("/data/", h.DeleteData)
	})

	return r, nil
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(res, req)
}
