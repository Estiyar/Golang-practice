package app

import (
	"fmt"
	"os"
	v1 "practice-7/internal/controller/http/v1"
	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/internal/usecase/repo"
	"practice-7/pkg/postgres"

	"github.com/gin-gonic/gin"
)

func Run() {
	pg, err := postgres.New()
	if err != nil {
		panic(fmt.Errorf("postgres: %w", err))
	}

	pg.Conn.AutoMigrate(&entity.User{})

	userRepo := repo.NewUserRepo(pg)
	userUseCase := usecase.NewUserUseCase(userRepo)

	handler := gin.Default()
	v1.NewRouter(handler, userUseCase)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8090"
	}
	handler.Run(":" + port)
}
