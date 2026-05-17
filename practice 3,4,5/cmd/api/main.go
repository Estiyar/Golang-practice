package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"practice3go/internal/handler"
	"practice3go/internal/middleware"
	"practice3go/internal/repository"
	"practice3go/internal/repository/_postgres"
	"practice3go/internal/usecase"
	"practice3go/pkg/modules"
)

func main() {
	cfg := &modules.PostgreConfig{
		Host:        getenv("DB_HOST", "localhost"),
		Port:        getenv("DB_PORT", "5432"),
		Username:    getenv("DB_USER", "postgres"),
		Password:    getenv("DB_PASSWORD", "postgres"),
		DBName:      getenv("DB_NAME", "mydb"),
		SSLMode:     getenv("DB_SSLMODE", "disable"),
		ExecTimeout: 5 * time.Second,
	}

	pg := _postgres.NewPGXDialect(context.Background(), cfg)
	repos := repository.NewRepositories(pg)
	uc := usecase.NewUserUsecase(repos.UserRepository)
	h := handler.NewUserHandler(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/users", h.Users)
	mux.HandleFunc("/users/", h.UserByID)
	mux.HandleFunc("/common-friends", h.CommonFriends)

	apiKey := getenv("API_KEY", "my-secret-key")
	finalHandler := middleware.Logging(middleware.APIKey(apiKey)(mux))

	port := getenv("APP_PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: finalHandler,
	}

	log.Println("listening on :" + port)
	log.Fatal(server.ListenAndServe())
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}