package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	 middle "github.com/mymindmap/api/internal/http/middleware"
	"github.com/mymindmap/api/internal/http/routes"
    "github.com/mymindmap/api/pkg/core/validator"
)

type Config struct {
	PostgresURL string
	JWTSecret   string
}

func loadConfig() (*Config, error) {
	_ = godotenv.Load()

	db := os.Getenv("POSTGRES_DB")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	jwtSecret := os.Getenv("JWT_SECRET")

	if db == "" || user == "" || pass == "" || host == "" {
		return nil, fmt.Errorf("missing database env vars")
	}

	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-change-in-production"
		//log.Println("WARNING: Using default JWT secret. Set JWT_SECRET env variable in production!")
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		user, pass, host, db,
	)

	return &Config{
		PostgresURL: connStr,
		JWTSecret:   jwtSecret,
	}, nil
}

func main() {
	// Загружаем конфиг
	conf, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	validator.Init()
	// Логирование в файл + stdout
	if err := os.MkdirAll("storage/logs", 0755); err == nil {
		logFilePath := filepath.Join("storage/logs", "server.log")
		if f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			log.SetOutput(f)
			defer f.Close()
		}
	}

	// Подключение к БД
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, conf.PostgresURL)
	if err != nil {
		log.Fatalf("db connection error: %v", err)
		log.Println("db connection error: %v", err)
	}
	defer dbpool.Close()


    // Роутер через DI
    mainRouter, err := routes.NewRouter(ctx, dbpool)
    if err != nil {
        log.Fatalf("router error: %v", err)
    }

    // Создаем chi роутер
    r := chi.NewRouter()

    // 🔒 SECURITY MIDDLEWARE - добавляем ПЕРВЫМ в цепочку
    securityConfig := middle.SecurityConfig{
        ExpectedHosts:    []string{"localhost", "127.0.0.1", "yourdomain.com"}, // укажите ваши домены
        EnableCSP:        true,
        EnableHSTS:       false, // true только для HTTPS в продакшене
        EnableFrameGuard: true,
    }
    r.Use(middle.SecurityMiddleware(securityConfig))
	r.Use(middle.JSONContentType)

    // Стандартные middleware chi (после security)
    r.Use(middleware.RequestID)    // Добавляет request_id в каждый запрос, для трейсинга
    r.Use(middleware.Logger)       // Логирование всех запросов  
    r.Use(middleware.Recoverer)    // Обработка паник
    r.Use(middleware.URLFormat)    // Парсер URLов

    // Подключаем основные маршруты
    r.Mount("/api", mainRouter)

	addr := ":8000"
	slog.Info("server started","address", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
