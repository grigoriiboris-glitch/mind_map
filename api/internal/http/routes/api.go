package routes

import (
	"context"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/mymindmap/api/internal/http/handlers"
	//"github.com/mymindmap/api/internal/services/log_service"
	//"github.com/mymindmap/api/repository"
	"github.com/mymindmap/api/internal/http/middleware"
)

func NewRouter(ctx context.Context, dbpool *pgxpool.Pool) (http.Handler, error) {
	r := chi.NewRouter()
	// ===== Repositories =====
	//logRepo := repository.NewLogRepository(dbpool)
	//userRepo := repository.NewUserRepository(dbpool)

	// ===== Handlers =====
	//logHandler := handlers.NewLogHandler(logService)

	r.Route("/", func(r chi.Router) {
		//Проверяем валидность  json и парсим его в тело запроса
		r.Use(middleware.JSONContentType)
		r.Use(middleware.AuthHeadersMiddleware)

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"mindmap_OK"}`))
        })
	})

	return r, nil
}