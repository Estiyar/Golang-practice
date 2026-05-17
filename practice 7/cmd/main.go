package main

import (
	"practice-7/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app.Run()
}
