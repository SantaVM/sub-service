package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sub-service/docs"
	"sub-service/internal/handler"
	mymw "sub-service/internal/handler/middleware"
	"sub-service/internal/infrastructure/config"
	"sub-service/internal/infrastructure/database"
	"sub-service/internal/infrastructure/logger"
	myv "sub-service/internal/infrastructure/validator"
	"sub-service/internal/repository"
	"sub-service/internal/service"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type App struct {
	server *http.Server
	db     *database.DBWrapper
	logger *slog.Logger
}

func New() (*App, error) {
	// config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// logger
	log := logger.New(logger.Config{
		Env: cfg.Env,
	})

	// database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// migrations
	if err := database.Migrate(db.DB); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	// validator
	v := myv.New()

	// DI (ручной)
	repo := repository.New(db.DB, log)
	svc := service.New(repo, log)
	h := handler.New(svc, log, v)

	// router
	router := newRouter(h, cfg, log)

	// server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		server: srv,
		db:     db,
		logger: log,
	}, nil
}

func (a *App) Run() error {
	// запуск сервера
	go func() {
		a.logger.Info("starting server", "addr", a.server.Addr)

		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return a.Close()
}

func (a *App) Close() error {
	a.logger.Info("closing resources")

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	return nil
}

func newRouter(h *handler.Handler, cfg *config.Config, log *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mymw.LoggerMw(log))
	r.Use(mymw.Recoverer(log))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// swagger
	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.BasePath = "/api/v1"

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// routes
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

	return r
}
