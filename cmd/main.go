package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"auto-pick-player-draft-format-fantasy-assignment/internal/db"
	"auto-pick-player-draft-format-fantasy-assignment/internal/handler"
	"auto-pick-player-draft-format-fantasy-assignment/internal/repository"
	"auto-pick-player-draft-format-fantasy-assignment/internal/service"
)

func main() {
	ctx := context.Background()

	// Load .env into the process env. Missing file is fine (e.g. env set by Docker/CI).
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set (see .env.example)")
	}
	port := getEnv("PORT", "8080")

	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	// Wire layers: repo -> service -> handler.
	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// getEnv returns the env var or a fallback if it's not set.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
