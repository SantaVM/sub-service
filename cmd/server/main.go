package main

import (
	"log"

	"sub-service/internal/app"
)

// @title Subscription Service API
// @version 1.0
// @description REST API для агрегации данных об онлайн подписках пользователей
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {

	// TODO: struct logging and chi logger
	// TODO: throttler? Timeout?
	// TODO: interfaces? structure?
	// TODO: refactor validation
	// TODO: tests

	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
