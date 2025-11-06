package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"container-manager/internal/server"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := server.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
