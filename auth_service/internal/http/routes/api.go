package routes

import (
	"context"
	"net/http"
	"log"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.etcd.io/bbolt"
	"github.com/mymindmap/api/internal/auth"
	"github.com/mymindmap/api/internal/http/handlers"
	//"github.com/mymindmap/api/internal/services/log_service"
	//"github.com/mymindmap/api/repository"
	"github.com/mymindmap/api/repositorybolt"
	"github.com/mymindmap/api/internal/http/middleware"
)

func NewRouter(ctx context.Context, dbpool *pgxpool.Pool, dbbolt *bbolt.DB , authConfig *auth.Config) (http.Handler, error) {
	r := chi.NewRouter()


	// Создаем репозиторий
	userRepo := repositorybolt.NewUserRepository(dbbolt)
	// Инициализируем buckets
	if err := userRepo.Initialize(); err != nil {
		log.Fatal(err)
	}
	// ===== Repositories =====
	//logRepo := repository.NewLogRepository(dbpool)
	//userRepo := repository.NewUserRepository(dbpool)

	// ===== Services =====
	//logService := log_service.NewLogService(logRepo)
	authService, err := auth.NewAuthService(userRepo, authConfig)

	if err != nil  { 
		log.Fatalf("auth service error: %v", err)
	}

	// ===== Handlers =====
	//logHandler := handlers.NewLogHandler(logService)
	auth_handler := handlers.NewAuthHandler(authService,userRepo,log.Default())

	r.Route("/", func(r chi.Router) {
		//Проверяем валидность  json и парсим его в тело запроса
		r.Use(middleware.JSONContentType)
		r.Route("/auth", func(r chi.Router) {
           r.Post("/login", auth_handler.Login)
           r.Post("/register", auth_handler.Register)
           r.Post("/logout", auth_handler.Logout)
           r.Post("/refresh", auth_handler.RefreshToken)
           r.Get("/check", middleware.AuthMiddleware(authService, auth_handler.Check))
           r.Get("/user", middleware.AuthMiddleware(authService, auth_handler.GetCurrentUser))
       })
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})

	return r, nil
}