package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sub-service/internal/config"
	"sub-service/internal/handler"
	"sub-service/internal/repository"
	"sub-service/internal/service"
	"sub-service/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	docs "sub-service/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Subscription Service API
// @version 1.0
// @description REST API для агрегации данных об онлайн подписках пользователей
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	// Инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Подключение к базе данных
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Выполнение миграций
	if err := database.Migrate(db); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Инициализация слоев
	repo := repository.New(db, logger)
	svc := service.New(repo, logger)
	h := handler.New(svc, logger)

	// Настройка роутера
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// TODO: struct logging and chi logger
	// TODO: DB closer
	// TODO: throttler? Timeout?
	// TODO: interfaces? structure?
	// TODO: tests

	// Swagger
	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Роуты
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/subscriptions", func(r chi.Router) {
			r.Post("/", h.CreateSubscription)
			r.Get("/", h.ListSubscriptions)
			r.Get("/{id}", h.GetSubscription)
			r.Put("/{id}", h.UpdateSubscription)
			r.Delete("/{id}", h.DeleteSubscription)
		})
		r.Get("/subscriptions/total", h.GetTotalCost)
		r.Get("/uuid", h.GetUUID)
	})

	// Запуск сервера
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		logger.Info("starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
